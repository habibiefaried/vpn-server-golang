package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/songgao/water"
)

func ReadIfaceAndSendTCP(ifce *water.Interface, conn net.Conn) {
	packet := make([]byte, MTU)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			log.Printf("Error reading from TUN interface: %v", err)
			n = 0
		}

		fmt.Printf("Read %d bytes from TUN interface\n", n)

		if n > 0 {
			// Send the data to the TCP server
			_, err = conn.Write(packet[:n])
			if err != nil {
				log.Printf("Error sending packet to TCP server: %v", err)
			}
		}
	}
}

func RecvTCPAndWriteIface(conn net.Conn, ifce *water.Interface) {
	response := make([]byte, 1500)
	for {
		n, err := conn.Read(response)
		if err != nil {
			log.Printf("Error reading from TCP server: %v", err)
			n = 0
		}

		fmt.Printf("Received %d bytes from TCP server\n", n)

		if n > 0 {
			// Write the response back to the TUN interface
			_, err = ifce.Write(response[:n])
			if err != nil {
				log.Printf("Error writing to TUN interface: %v", err)
			}
		}
	}
}

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Running command error: ", err)
	}
	log.Printf("%s\n", stdoutStderr)
}
