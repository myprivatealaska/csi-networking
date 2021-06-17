package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"strings"

	"github.com/google/uuid"
)

type DnsType uint16

const (
	// QTypeA a host address
	QTypeA DnsType = 1
	// QTypeNS an authoritative name server
	QTypeNS    DnsType = 2
	QTypeCNAME DnsType = 5
	QTypeMX    DnsType = 15
)

type DnsClass uint16

const ClassInternet DnsClass = 1

type DnsHeader struct {
	ID      uint16
	Flags   uint16
	NumQ    uint16
	NumA    uint16
	NumRR   uint16
	NumExRR uint16
}

type DnsRR struct {
	Name     string
	Type     DnsType
	Class    DnsClass
	TTL      uint32
	RdLength uint16
	RData    string
}

type DnsQuestion struct {
	Name  string
	Type  DnsType
	Class DnsClass
}

type DnsAnswer []DnsRR

type DnsMessage struct {
	Header         DnsHeader
	Questions      []DnsQuestion
	Answers        []DnsAnswer
	Authority      []DnsRR
	AdditionalInfo []DnsRR
}

func NewDnsResolveHostQuestionMessage(host string) *DnsMessage {
	question := DnsQuestion{
		Name:  host,
		Type:  QTypeA,
		Class: ClassInternet,
	}
	header := DnsHeader{
		ID:      uint16(uuid.New().ID()),
		Flags:   uint16(0b0000000100000000),
		NumQ:    1,
		NumA:    0,
		NumRR:   0,
		NumExRR: 0,
	}
	log.Println(header.ID)
	return &DnsMessage{
		Header:    header,
		Questions: []DnsQuestion{question},
	}
}

func (m *DnsMessage) Serialize() []byte {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, m.Header)
	checkErr(err)

	byteArr := buf.Bytes()

	for _, q := range m.Questions {
		for _, p := range strings.Split(q.Name, ".") {
			byteArr = append(byteArr, byte(len(p)))
			byteArr = append(byteArr, []byte(p)...)
		}
		byteArr = append(byteArr, 0x00)
		byteArr = append(byteArr, []byte{0, 0, 0, 0}...)

		tBytes := make([]byte, 8)
		binary.BigEndian.PutUint16(tBytes, uint16(q.Type))

		cBytes := make([]byte, 8)
		binary.BigEndian.PutUint16(cBytes, uint16(q.Class))

		byteArr = append(byteArr, tBytes...)
		byteArr = append(byteArr, cBytes...)
	}

	return byteArr
}

func DeserealizeDnsResponse(message []byte) {
	var header DnsHeader
	err := binary.Read(bytes.NewReader(message[:96]), binary.BigEndian, &header)
	checkErr(err)

	log.Println(header.NumQ)
}
