package main

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (PI *PooledIface) ReadInterfaceandDistribute() {
	rp := make([]byte, BufferSize)
	for {
		// continuously reading interface
		n, err := PI.Ifce.Read(rp)
		if err != nil {
			log.Printf("Error reading from TUN interface: %v", err)
			n = 0
		}

		if n > 0 {
			packet := gopacket.NewPacket(rp[:n], layers.LayerTypeIPv4, gopacket.Default)

			if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv4)

				fmt.Printf("IPv4 packet: Src: %s, Dest: %s\n", ip.SrcIP, ip.DstIP)
				if val, ok := PI.ConnectionsPool[ip.DstIP.String()]; ok {
					val.SendTCPMessage(rp[:n])
				} else {
					log.Printf("Destination %v not existent\n", ip.DstIP)
				}
			} else {
				fmt.Println("Not an IPv4 packet")
			}
		}
	}
}
