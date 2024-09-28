package main

import (
	"log"
	"os/exec"
)

func runIP(args ...string) {
	cmd := exec.Command("/sbin/ip", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Running command error: ", err)
	}
	log.Printf("%s\n", stdoutStderr)
}
