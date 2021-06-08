package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"sort"
)

const (
	PacketHeaderStart           int64 = 24
	PacketSizeOffset            int64 = 8
	PacketUntruncatedSizeOffset int64 = 12
	PacketStartOffset           int64 = 16
)

type HttpPacket struct {
	FromServer     bool
	Data           []byte
	SequenceNumber uint32
}

type BySequenceNumber []HttpPacket

func (p BySequenceNumber) Len() int           { return len(p) }
func (p BySequenceNumber) Less(i, j int) bool { return p[i].SequenceNumber < p[j].SequenceNumber }
func (p BySequenceNumber) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {

	currDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	f, err := os.Open(fmt.Sprintf("%v/01-layers-of-protocols/net.cap", currDir))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var payloadHeaderStart int64 = 16
	payloadSizeBytes := make([]byte, 4)
	_, err = f.ReadAt(payloadSizeBytes, payloadHeaderStart)
	check(err)

	payloadSize := byteArrToIntLittleEndian(payloadSizeBytes)
	fmt.Printf("%d\n", payloadSize)

	httpPackets := map[uint32]HttpPacket{}
	packetCount := 0
	var packetHeaderStart = PacketHeaderStart

	for {
		packetSizeBytes := make([]byte, 4)
		_, err = f.ReadAt(packetSizeBytes, packetHeaderStart+PacketSizeOffset)
		if err == io.EOF {
			break
		} else {
			check(err)
		}
		packetSize := byteArrToIntLittleEndian(packetSizeBytes)

		untruncatedPacketSizeBytes := make([]byte, 4)
		_, err = f.ReadAt(untruncatedPacketSizeBytes, packetHeaderStart+PacketUntruncatedSizeOffset)
		if err == io.EOF {
			break
		} else {
			check(err)
		}
		untruncatedPacketSize := byteArrToIntLittleEndian(untruncatedPacketSizeBytes)

		fmt.Printf("Packet size: %d Untruncated packet size: %d\n", packetSize, untruncatedPacketSize)

		fmt.Printf("Packet header start: %d\n", packetHeaderStart)

		packetDataBytes := make([]byte, packetSize)
		_, err = f.ReadAt(packetDataBytes, packetHeaderStart+PacketStartOffset)
		if err == io.EOF {
			break
		} else {
			check(err)
		}
		httpPacket := parseEthernet(packetDataBytes)
		if httpPacket.FromServer {
			httpPackets[httpPacket.SequenceNumber] = httpPacket
		}
		packetCount++
		packetHeaderStart = packetHeaderStart + PacketStartOffset + packetSize
	}

	fmt.Printf("Packet count: %d\n", packetCount)

	httpPacketArray := make([]HttpPacket, 0)
	for _, v := range httpPackets {
		httpPacketArray = append(httpPacketArray, v)
	}

	sort.Sort(BySequenceNumber(httpPacketArray))

	values := make([][]byte, packetCount)
	for i, packet := range httpPacketArray {
		log.Printf("Seq Number %d\n", packet.SequenceNumber)
		values[i] = packet.Data
	}

	// Combine fragments
	response := bytes.Join(values, []byte{})

	// Split into HTTP header and body
	parts := bytes.SplitN(response, []byte{'\r', '\n', '\r', '\n'}, 2)

	fmt.Printf("Http headers: %v\n", string(parts[0]))

	imageFile, err := os.Create(fmt.Sprintf("%v/01-layers-of-protocols/img.jpg", currDir))
	check(err)
	defer imageFile.Close()

	_, err = imageFile.Write(parts[1])

	check(err)
}

func parseEthernet(data []byte) HttpPacket {
	destMac := data[:6]
	fmt.Printf("Destination MAC address: %x\n", destMac)

	sourceMac := data[6:12]
	fmt.Printf("Source MAC address: %x\n", sourceMac)

	etherType := data[12:14]
	fmt.Printf("Ether type: %x\n", etherType)

	payload := data[14:]
	return parseIPv4(payload)
}

func parseIPv4(data []byte) HttpPacket {
	intHeaderLengthBytes := (data[0] & 0x0f) * 4
	fmt.Printf("Internet Header Length, bytes : %d\n", intHeaderLengthBytes)

	totalLengthBytes := data[2:4]
	fmt.Printf("Total payload length: %d\n", bytesToInt(totalLengthBytes))

	protocolBytes := data[9]
	fmt.Printf("Protocol: %x\n", protocolBytes)

	sourceIPAddr := data[12:16]
	destIPAddr := data[16:20]
	fmt.Printf("Source IP: %v, Destination IP: %v\n", sourceIPAddr, destIPAddr)

	return parseTCP(data[intHeaderLengthBytes:])
}

func parseTCP(data []byte) HttpPacket {
	sourcePortBytes := data[0:2]
	destPortBytes := data[2:4]

	flags := bytesToInt(data[12:14])
	isSYN := (flags & (1 << 1)) > 0

	fmt.Printf("Source Port: %d Destination Port: %d\n", bytesToInt(sourcePortBytes), bytesToInt(destPortBytes))

	seqNumberBytes := data[4:8]
	fmt.Printf("Sequence Number: %d\n", bytesToInt(seqNumberBytes))

	headerLengthBytes := (data[12] >> 4) * 4
	fmt.Printf("TCP Header length: %d\n", headerLengthBytes)

	httpData := data[headerLengthBytes:]
	return HttpPacket{
		Data:           httpData,
		FromServer:     bytesToInt(sourcePortBytes) == 80 && !isSYN,
		SequenceNumber: binary.BigEndian.Uint32(seqNumberBytes),
	}
}

func bytesToInt(bytes []byte) int64 {
	z := new(big.Int)
	return z.SetBytes(bytes).Int64()
}

func byteArrToIntLittleEndian(bytes []byte) int64 {
	l := len(bytes)
	var value int64
	value |= int64(bytes[0])
	if l > 1 {
		value |= int64(bytes[1]) << 8
	}
	if l > 2 {
		value |= int64(bytes[2]) << 16
	}
	if l > 3 {
		value |= int64(bytes[3]) << 24
	}
	return value
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
