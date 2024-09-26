package main

import (
	"io"
	"log"

	"github.com/vishvananda/netlink"
)

// transfer data between two interfaces (TUN or TCP)
func transfer(src io.Reader, dst io.Writer) {
	buf := make([]byte, 1024) // Larger buffer
	for {
		n, err := src.Read(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}
		log.Printf("Read %d bytes", n)

		_, err = dst.Write(buf[:n])
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
		log.Printf("Wrote %d bytes", n)
	}
}

// Setup the interface using the netlink package
func setupInterface(ifaceName string, ipCIDR string) error {
	// Find the link for the TUN interface
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return err
	}

	// Set the MTU of the TUN interface
	if err := netlink.LinkSetMTU(link, 1300); err != nil {
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
