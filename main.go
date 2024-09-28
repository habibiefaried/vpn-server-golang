package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
	"github.com/spf13/pflag"
)

const (
	MTU        = 1300
	BufferSize = 16 * 1024
)

func main() {
	var host *string = pflag.String("host", "0.0.0.0", "Which host/IP this VPN is using")
	var port *string = pflag.String("port", "10443", "Which port this VPN is using")
	var isProxyTypeClient *bool = pflag.Bool("isclient", true, "Is this client side or server side")
	var netIp *string = pflag.String("netip", "", "VPN IP that will be advertised (x.x.x.x/y) format")
	// var EncKey *string = flag.String("key", "AbcD1234!_D3f4ult", "Encryption key is being used")
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
		// This is VPN client hence needs to be connected to VPN server
		ifc := IfaceConn{
			Ifce: ifce,
		}

		ifc.DialUp(fmt.Sprintf("%v:%v", *host, *port))
		IpParts := strings.Split(*netIp, "/")
		ifc.SendTCPMessage(IpParts[0])
		go ifc.ReadIfaceAndSendTCP()
		ifc.RecvTCPAndWriteIface()

	} else {
		connectionsPool := make(map[string]IfaceConn)
		listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", *host, *port))
		if err != nil {
			log.Fatalf("Failed to listen on port %v: %v", *port, err)
		}
		log.Printf("Listening on port %v\n", *port)

		go func() {
			rp := make([]byte, BufferSize)
			for {
				// continuously reading interface
				n, err := ifce.Read(rp)
				if err != nil {
					log.Printf("Error reading from TUN interface: %v", err)
					n = 0
				}

				if n > 0 {
					packet := gopacket.NewPacket(rp[:n], layers.LayerTypeIPv4, gopacket.Default)

					if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
						ip, _ := ipLayer.(*layers.IPv4)

						fmt.Printf("IPv4 packet: Src: %s, Dest: %s\n", ip.SrcIP, ip.DstIP)
						// Send the data to the TCP server
						// TODO: Add retry / timeout
						if val, ok := connectionsPool[ip.DstIP.String()]; ok {
							_, err = val.Conn.Write(rp[:n])
							if err != nil {
								log.Printf("Error sending packet to TCP server: %v", err)
							}
						} else {
							log.Printf("Destination %v not existent\n", ip.DstIP)
						}
					} else {
						fmt.Println("Not an IPv4 packet")
					}
				}
			}
		}()

		// Accept client connections in a loop
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
				ifc := IfaceConn{
					Ifce: ifce,
					Conn: conn,
				}
				netIp := ifc.RecvTCPMessage()
				log.Println("Got IP from " + netIp)

				connectionsPool[netIp] = ifc

				ifc.RecvTCPAndWriteIface()
			}()
		}
	}

}
