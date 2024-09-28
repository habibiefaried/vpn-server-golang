package main

import (
	"net"

	"github.com/songgao/water"
)

// IfaceConn stores net connection, interface
type IfaceConn struct {
	Conn net.Conn
	Ifce *water.Interface
}
