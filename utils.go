package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
)

func ReadIfaceAndSendTCP(ifce *water.Interface, conn net.Conn) {
	rp := make([]byte, MTU)
	for {
		n, err := ifce.Read(rp)
		if err != nil {
			log.Printf("Error reading from TUN interface: %v", err)
			n = 0
		}

		packet := gopacket.NewPacket(rp[:n], layers.LayerTypeIPv4, gopacket.Default)

		// Check if the packet contains an IPv4 layer
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)

			fmt.Printf("IPv4 packet: Src: %s, Dest: %s\n", ip.SrcIP, ip.DstIP)
		} else {
			fmt.Println("Not an IPv4 packet")
		}

		if n > 0 {
			// Send the data to the TCP server
			// TODO: Add retry or kill
			_, err = conn.Write(rp[:n])
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
			// TODO: Add retry or kill
			n = 0
		}

		if n > 0 {
			fmt.Printf("Received %d bytes from TCP server\n", n)
			// Write the response back to the TUN interface
			_, err = ifce.Write(response[:n])
			if err != nil {
				log.Printf("Error writing to TUN interface: %v", err)
			}
		}
	}
}

func sendTCPMessage(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
}

func recvTCPMessage(conn net.Conn) string {
	// Read the response from the server
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}
	return string(buffer[:n])
}

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Running command error: ", err)
	}
	log.Printf("%s\n", stdoutStderr)
}
