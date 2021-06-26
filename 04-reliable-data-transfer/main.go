package main

import (
	"strings"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

const (
	Proxy1EntryPort = 58627
	SenderPort      = 52059
	ReceiverPort    = 52058
	Proxy2EntryPort = 62489
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go receiver(ReceiverPort, Proxy2EntryPort)
	go sender(SenderPort, Proxy1EntryPort)
	wg.Wait()
}

func sender(port int, sendToPort int) {
	c := NewReliableUdpClient(port, sendToPort)
	defer syscall.Close(c.Socket)
	c.SendDatagram([]byte{'h', 'e', 'y'})
	c.SendDatagram([]byte{'y', 'o', 'u'})
	c.SendDatagram([]byte{'r'})
}

func receiver(port int, sendToPort int) {
	c := NewReliableUdpClient(port, sendToPort)
	defer syscall.Close(c.Socket)

	for {
		received, err := c.Receive()
		if err != nil {
			if !strings.Contains(err.Error(), "Response segment") {
				checkErr(err)
			}
			c.SendDatagram([]byte{'N', 'A', 'K'})
		} else {
			c.Logger.Info("Checksum", zap.Any("Checksum", received.Checksum))
			c.SendDatagram([]byte{'A', 'C', 'K'})
		}
	}
}
