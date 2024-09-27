package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
	"github.com/spf13/pflag"
)

const (
	MTU = 1300
)

func main() {
	var host *string = pflag.String("host", "0.0.0.0", "Which host/IP this VPN is using")
	var port *string = pflag.String("port", "10443", "Which port this VPN is using")
	var isProxyTypeClient *bool = pflag.Bool("isclient", true, "Is this client side or server side")
	var netIp *string = pflag.String("netip", "", "VPN IP that will be advertised (x.x.x.x/y) format")
	pflag.Parse()

	// Setup the TUN interface
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatalf("Unable to allocate TUN interface: %v", err)
	}
	log.Printf("Interface allocated: %s\n", ifce.Name())

	runIP("link", "set", "dev", ifce.Name(), "mtu", fmt.Sprintf("%v", MTU))
	runIP("addr", "add", *netIp, "dev", ifce.Name())
	runIP("link", "set", "dev", ifce.Name(), "up")

	if *isProxyTypeClient {
		// Connect to the VPN server
		connAttempts := 3

		var conn net.Conn
		var err error
		for i := 0; i < connAttempts; i++ {
			conn, err = net.Dial("tcp", fmt.Sprintf("%v:%v", *host, *port)) // Replace with the actual server IP
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
		log.Println("Connected to server")

		// Handle bidirectional communication
		IpParts := strings.Split(*netIp, "/")
		sendTCPMessage(conn, IpParts[0])
		go ReadIfaceAndSendTCP(ifce, conn)
		RecvTCPAndWriteIface(conn, ifce)

	} else {
		connectionsPool := make(map[string]net.Conn)

		// Accept client connections in a loop
		listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", *host, *port))
		if err != nil {
			log.Fatalf("Failed to listen on port %v: %v", *port, err)
		}
		log.Printf("Listening on port %v\n", *port)

		go func() {
			rp := make([]byte, MTU)
			for {
				n, err := ifce.Read(rp) // TODO: THIS MUST MOVE OUT FROM HERE IF WE WANT MULTIPLE USER
				if err != nil {
					log.Printf("Error reading from TUN interface: %v", err)
					n = 0
				}

				if n > 0 {
					packet := gopacket.NewPacket(rp[:n], layers.LayerTypeIPv4, gopacket.Default)

					// Check if the packet contains an IPv4 layer
					if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
						ip, _ := ipLayer.(*layers.IPv4)

						fmt.Printf("IPv4 packet: Src: %s, Dest: %s\n", ip.SrcIP, ip.DstIP)
						// Send the data to the TCP server
						// TODO: Add retry / timeout
						_, err = connectionsPool[ip.DstIP.String()].Write(rp[:n])
						if err != nil {
							log.Printf("Error sending packet to TCP server: %v", err)
						}
					} else {
						fmt.Println("Not an IPv4 packet")
					}
				}
			}
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept client: %v", err)
				continue
			}
			log.Println("Client connected")

			// Handle each client connection in a new goroutine
			go func() {
				defer conn.Close()
				netIp := recvTCPMessage(conn)
				log.Println("Got IP from " + netIp)

				connectionsPool[netIp] = conn

				RecvTCPAndWriteIface(connectionsPool[netIp], ifce)
			}()
		}
	}

}
