package main

import (
	"net"

	"github.com/songgao/water"
)

// IfaceConn stores net connection and corresponding interface
type IfaceConn struct {
	Conn net.Conn
	Ifce *water.Interface
}

// PooledIface stores all IfaceConn to one interface
// Main purpose is to read, select and write to corresponding packet
type PooledIface struct {
	ConnectionsPool map[string]IfaceConn
	Ifce            *water.Interface
}
