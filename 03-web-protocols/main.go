package main

import (
	"log"
	"syscall"

	"github.com/pkg/errors"
)

const (
	ProxyPort  = 8001
	ServerPort = 9000
)

func main() {
	// Open a socket for proxy server to communicate with clients
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	checkErr(errors.Wrap(err, "Can't open a socket"))
	defer syscall.Close(fd)

	proxyAddr := &syscall.SockaddrInet4{Port: ProxyPort}
	err = syscall.Bind(fd, proxyAddr)
	checkErr(errors.Wrap(err, "Can't bind to proxy port"))

	err = syscall.Listen(fd, 100)
	checkErr(errors.Wrap(err, "Can't start listening on socket"))

	var nfd int
	defer syscall.Close(nfd)

	for {
		receiveBuf := make([]byte, 4096)
		// Get ready to accept incoming connections on the proxy <> clients socket
		nfd, _, err = syscall.Accept(fd)
		checkErr(errors.Wrap(err, "Can't accept on the socket"))

		// Receive incoming connections on the socket
		n, _, err := syscall.Recvfrom(nfd, receiveBuf, 0)
		checkErr(errors.Wrap(err, "Can't receive on the socket"))

		respBytes := forwardReq(receiveBuf[:n])

		log.Println(string(respBytes))

		err = syscall.Sendto(nfd, respBytes, 0, proxyAddr)
		checkErr(errors.Wrap(err, "Can't send the response on the socket"))
	}
}

func forwardReq(reqBytes []byte) []byte {
	// Open a socket for the proxy <> server communication. Proxy is a TCP client of the server.
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	checkErr(errors.Wrap(err, "Can't open a socket"))

	defer syscall.Close(fd)

	serverAddr := &syscall.SockaddrInet4{
		Port: ServerPort,
	}
	// Connect to server's TCP socket
	err = syscall.Connect(fd, serverAddr)
	checkErr(errors.Wrap(err, "Can't connect to the server's socket"))

	err = syscall.Sendto(fd, reqBytes, 0, serverAddr)
	checkErr(errors.Wrap(err, "Can't send data to the server's socket"))

	var receiveBuf = make([]byte, 4096)
	n, _, err := syscall.Recvfrom(fd, receiveBuf, 0)
	checkErr(errors.Wrap(err, "Can't receive data on the server's socket"))

	return receiveBuf[:n]
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
