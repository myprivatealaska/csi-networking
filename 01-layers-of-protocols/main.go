package main

import (
	"fmt"
	"io"
	"os"
)

const (
	PacketHeaderStart           int64 = 24
	PacketSizeOffset            int64 = 8
	PacketUntruncatedSizeOffset int64 = 12
	PacketStartOffset           int64 = 16
)

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

	payloadSize := byteArrToInt(payloadSizeBytes)
	fmt.Printf("%d\n", payloadSize)

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
		packetSize := byteArrToInt(packetSizeBytes)

		untruncatedPacketSizeBytes := make([]byte, 4)
		_, err = f.ReadAt(untruncatedPacketSizeBytes, packetHeaderStart+PacketUntruncatedSizeOffset)
		if err == io.EOF {
			break
		} else {
			check(err)
		}
		untruncatedPacketSize := byteArrToInt(untruncatedPacketSizeBytes)

		fmt.Printf("Packet size: %d Untruncated packet size: %d\n", packetSize, untruncatedPacketSize)

		packetCount++
		packetHeaderStart = packetHeaderStart + PacketStartOffset + packetSize

		fmt.Printf("Packet header start: %d\n", packetHeaderStart)
	}

	fmt.Printf("Packet count: %d\n", packetCount)
}

func byteArrToInt(bytes []byte) int64 {
	var value int64
	value |= int64(bytes[0])
	value |= int64(bytes[1]) << 8
	value |= int64(bytes[2]) << 16
	value |= int64(bytes[3]) << 24

	return value
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
