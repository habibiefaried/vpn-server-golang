package main

import (
	"io"
	"log"
	"os/exec"
)

// transfer data between two interfaces (TUN or TCP)
func transfer(src io.Reader, dst io.Writer) {
	buf := make([]byte, 2000) // Larger buffer
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

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Running command error: ", err)
	}
	log.Printf("%s\n", stdoutStderr)
}
