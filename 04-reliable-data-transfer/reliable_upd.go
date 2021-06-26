package main

import (
	"bytes"
	"encoding/gob"
	"strings"
	"syscall"

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
	Checksum        uint16
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

	segmBytes := buf.Bytes()
	check := checksum(segmBytes)

	segment := Segment{
		SourcePort:      raw.SourcePort,
		DestinationPort: raw.DestinationPort,
		DataLength:      raw.DataLength,
		Checksum:        check,
		Data:            raw.Data,
	}

	var newBuf bytes.Buffer
	gFinal := gob.NewEncoder(&newBuf)

	err = gFinal.Encode(segment)
	checkErr(err)

	err = syscall.Sendto(c.Socket, newBuf.Bytes(), 0, &syscall.SockaddrInet4{Port: c.DestinationPort})
	checkErr(err)

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

	receivedCheck := checksum(receiveBuf[:n])
	if receivedCheck != 0b1111111111111111 {
		return respSeg, errors.Wrap(err, "Response segment checksum err")
	}

	c.Logger.Info("Received a segment", zap.Any("Segment", string(respSeg.Data)))

	return respSeg, nil
}

// Internet checksum
// https://datatracker.ietf.org/doc/html/rfc1071.html
func checksum(buf []byte) uint16 {
	sum := uint32(0)

	// (1)  Adjacent octets to be checksummed are paired to form 16-bit
	// integers, and the 1's complement sum of these 16-bit integers is
	// formed.
	for ; len(buf) >= 2; buf = buf[2:] {
		sum += uint32(buf[0])<<8 | uint32(buf[1])
	}

	// If the total length is odd, the received data is padded with one
	// octet of zeros for computing the checksum.  This checksum may be
	// replaced in the future.
	if len(buf) > 0 {
		sum += uint32(buf[0]) << 8
	}

	// On a 2's complement machine, the 1's complement sum must be
	// computed by means of an "end around carry", i.e., any overflows
	// from the most significant bits are added into the least
	// significant bits. See the examples below.
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}

	csum := ^uint16(sum)
	return csum
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
