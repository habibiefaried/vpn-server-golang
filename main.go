package main

import (
	"fmt"
	"log"
	"net"
	"strings"

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
		ifc.SendTCPMessage([]byte(IpParts[0]))
		go ifc.ReadIfaceAndSendTCP()
		ifc.RecvTCPAndWriteIface()

	} else {
		listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", *host, *port))
		if err != nil {
			log.Fatalf("Failed to listen on port %v: %v", *port, err)
		}
		log.Printf("Listening on port %v\n", *port)

		PI := PooledIface{
			Ifce: ifce,
		}
		PI.ConnectionsPool = make(map[string]IfaceConn)

		go PI.ReadInterfaceandDistribute()

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
				ifc := IfaceConn{
					Ifce: ifce,
					Conn: conn,
				}
				netIp := string(ifc.RecvTCPMessage())
				log.Println("Got IP from " + netIp)

				PI.ConnectionsPool[netIp] = ifc

				ifc.RecvTCPAndWriteIface()
			}()
		}
	}

}
