package main

import (
	"syscall"
)

const (
	DNSPort int = 53
)

func main() {
	// Open a socket
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	checkErr(err)

	// Bind to a port
	err = syscall.Bind(socket, &syscall.SockaddrInet4{})
	checkErr(err)

	dnsMessage := NewDnsResolveHostQuestionMessage("topresume.com")
	messageBytes := dnsMessage.Serialize()

	googleDnsAddr := &syscall.SockaddrInet4{
		Port: DNSPort,
		Addr: [4]byte{8, 8, 8, 8},
	}
	err = syscall.Sendto(socket, messageBytes, 0, googleDnsAddr)
	checkErr(err)

	recvBuf := make([]byte, 4096)

	for {
		_, _, err = syscall.Recvfrom(socket, recvBuf, 0)
		checkErr(err)

		// Deserialize
		//err = binary.Read(bytes.NewReader(recvBuf), binary.BigEndian, dnsAnswer)
		//checkErr(err)

		DeserealizeDnsResponse(recvBuf)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
