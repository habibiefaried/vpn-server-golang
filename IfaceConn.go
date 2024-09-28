package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (ifc *IfaceConn) DialUp(dest string) {
	connAttempts := 3
	var err error

	for i := 0; i < connAttempts; i++ {
		ifc.Conn, err = net.Dial("tcp", dest)
		if err == nil {
			log.Println("Connected to server")
			break
		}
		log.Printf("Failed to connect to server (attempt %d/%d): %v", i+1, connAttempts, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to server after %d attempts: %v", connAttempts, err)
	}
}

func (ifc *IfaceConn) ReadIfaceAndSendTCP() {
	rp := make([]byte, BufferSize)
	for {
		n, err := ifc.Ifce.Read(rp)
		if err != nil {
			log.Printf("Error reading from TUN interface: %v", err)
			n = 0
		}

		packet := gopacket.NewPacket(rp[:n], layers.LayerTypeIPv4, gopacket.Default)

		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)

			fmt.Printf("IPv4 packet: Src: %s, Dest: %s\n", ip.SrcIP, ip.DstIP)
		} else {
			fmt.Println("Not an IPv4 packet")
		}

		if n > 0 {
			// Send the data to the TCP server
			// TODO: Add retry or kill
			_, err = ifc.Conn.Write(rp[:n])
			if err != nil {
				log.Printf("Error sending packet to TCP server: %v", err)
			}
		}
	}
}

func (ifc *IfaceConn) RecvTCPAndWriteIface() {
	response := make([]byte, BufferSize)
	for {
		n, err := ifc.Conn.Read(response)
		if err != nil {
			// TODO: Add retry or kill
			n = 0
		}

		if n > 0 {
			fmt.Printf("Received %d bytes from TCP server\n", n)
			// Write the response back to the TUN interface
			_, err = ifc.Ifce.Write(response[:n])
			if err != nil {
				log.Printf("Error writing to TUN interface: %v", err)
			}
		}
	}
}

func (ifc *IfaceConn) SendTCPMessage(message string) {
	_, err := ifc.Conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
}

func (ifc *IfaceConn) RecvTCPMessage() string {
	buffer := make([]byte, BufferSize)
	n, err := ifc.Conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}
	return string(buffer[:n])
}
