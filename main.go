package main

import (
	"log"
	"os"
	"syscall"
)

func main() {
	// 向当前进程发送 SIGHUP
	err := syscall.Kill(os.Getpid(), syscall.SIGHUP)
	if err != nil {
		log.Fatal("Failed to send SIGHUP:", err)
	}
}
