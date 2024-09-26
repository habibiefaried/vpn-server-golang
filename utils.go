package main

import (
	"io"
	"log"
	"net"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// handleClient handles bidirectional communication between the server and the client
func handleClient(conn net.Conn, ifce *water.Interface) {
	defer conn.Close()

	// Transfer data between the TUN interface and the client
	go transfer(conn, ifce)
	transfer(ifce, conn)
}

// transfer data between two interfaces (TUN or TCP)
func transfer(dst io.Writer, src io.Reader) {
	buf := make([]byte, 1500)
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed")
				break
			}
			log.Println("Read error:", err)
			return
		}
		_, err = dst.Write(buf[:n])
		if err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

// Setup the interface using the netlink package
func setupInterface(ifaceName string, ipCIDR string) error {
	// Find the link for the TUN interface
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return err
	}

	// Parse the IP address and CIDR
	addr, err := netlink.ParseAddr(ipCIDR)
	if err != nil {
		return err
	}

	// Assign the IP address to the TUN interface
	if err := netlink.AddrAdd(link, addr); err != nil {
		return err
	}

	// Bring the TUN interface up
	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	log.Printf("Interface %s set up with IP %s\n", ifaceName, ipCIDR)
	return nil
}
