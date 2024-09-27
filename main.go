package main

import (
	"fmt"
	"log"
	"net"
	"time"

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
		go ReadIfaceAndSendTCP(ifce, conn)
		RecvTCPAndWriteIface(conn, ifce)

	} else {
		// Accept client connections in a loop
		listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", *host, *port))
		if err != nil {
			log.Fatalf("Failed to listen on port %v: %v", *port, err)
		}
		log.Printf("Listening on port %v\n", *port)

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
				go RecvTCPAndWriteIface(conn, ifce)
				ReadIfaceAndSendTCP(ifce, conn)
			}()
		}
	}

}
