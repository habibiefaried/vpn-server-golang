package main

import (
	"io"
	"log"
	"os/exec"
)

// transfer data between two interfaces (TUN or TCP)
func transfer(src io.Reader, dst io.Writer) {
	n, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("Transfer error: %v", err)
	}
	log.Printf("Transferred %d bytes", n)
}

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Running command error: ", err)
	}
	log.Printf("%s\n", stdoutStderr)
}
