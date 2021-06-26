package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ReliableUdpClient struct {
	Socket          int
	Port            int
	DestinationPort int
	Logger          zap.Logger
}

type Segment struct {
	SourcePort      uint16
	DestinationPort uint16
	DataLength      uint16
	Checksum        [32]byte
	Data            []byte
}

type RawSegment struct {
	SourcePort      uint16
	DestinationPort uint16
	DataLength      uint16
	Data            []byte
}

func NewReliableUdpClient(port int, destPort int) *ReliableUdpClient {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	checkErr(errors.Wrap(err, "Can't open a socket"))

	err = syscall.Bind(fd, &syscall.SockaddrInet4{Port: port})
	checkErr(errors.Wrap(err, "Can't bind to a port"))

	l, err := zap.NewDevelopment()
	checkErr(errors.Wrap(err, "Can't create logger"))

	c := &ReliableUdpClient{
		Socket: fd,
		Port:   port,
		Logger: *l,
	}

	if destPort != 0 {
		c.DestinationPort = destPort
	}

	return c
}

func (c *ReliableUdpClient) SendDatagram(data []byte) {
	// prepare bytes, send them to the socket
	var buf bytes.Buffer
	g := gob.NewEncoder(&buf)

	raw := RawSegment{
		SourcePort:      uint16(c.Port),
		DestinationPort: uint16(c.DestinationPort),
		DataLength:      uint16(len(data)),
		Data:            data,
	}

	err := g.Encode(raw)
	checkErr(err)

	segment := Segment{
		SourcePort:      uint16(c.Port),
		DestinationPort: uint16(c.DestinationPort),
		DataLength:      uint16(len(data)),
		Checksum:        sha256.Sum256(buf.Bytes()),
		Data:            data,
	}

	var newBuf bytes.Buffer
	gFinal := gob.NewEncoder(&newBuf)

	err = gFinal.Encode(segment)
	checkErr(err)

	err = syscall.Sendto(c.Socket, newBuf.Bytes(), 0, &syscall.SockaddrInet4{Port: c.DestinationPort})
	checkErr(err)

	time.Sleep(3 * time.Second)

	// wait for ACK / NAK, retransmit if necessary
	ackSeg, err := c.Receive()
	if err != nil {
		if !strings.Contains(err.Error(), "Response segment decoding err") {
			checkErr(err)
		}
		c.Logger.Info("Error reading ACK/NAK. Gotta retransmit")
	} else {
		c.Logger.Info("Received ACK/NAK", zap.String("ACK/NAK Segment", string(ackSeg.Data)))
	}
}

func (c *ReliableUdpClient) Receive() (Segment, error) {
	var receiveBuf = make([]byte, 1024)
	var respSeg Segment

	n, _, err := syscall.Recvfrom(c.Socket, receiveBuf, 0)
	if err != nil {
		return respSeg, err
	}

	dec := gob.NewDecoder(bytes.NewReader(receiveBuf[:n]))

	err = dec.Decode(&respSeg)
	if err != nil {
		return respSeg, errors.Wrap(err, "Response segment decoding err")
	}

	// Check checksum
	raw := RawSegment{
		SourcePort:      respSeg.SourcePort,
		DestinationPort: respSeg.DestinationPort,
		DataLength:      respSeg.DataLength,
		Data:            respSeg.Data,
	}

	var buf bytes.Buffer
	g := gob.NewEncoder(&buf)

	err = g.Encode(raw)
	checkErr(err)

	check := sha256.Sum256(buf.Bytes())

	if respSeg.Checksum != check {
		return respSeg, errors.Wrap(err, "Response segment checksum err")
	}

	c.Logger.Info("Received a segment", zap.Any("Segment", string(respSeg.Data)))

	return respSeg, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
