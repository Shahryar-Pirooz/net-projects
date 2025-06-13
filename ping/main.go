package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	if len(os.Args) <= 1 {
		errorHandler("Usage error", "Usage: ping <IP address>")
	}
	ip := os.Args[1]
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		errorHandler("Failed to listen for ICMP packets", err.Error())
	}
	defer conn.Close()

	dst, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		errorHandler("Failed to resolve IP address", err.Error())
	}
	seq := 0
	for {
		seq++
		start := time.Now()
		echo := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  seq,
				Data: []byte("hello server"),
			},
		}
		marEcho, err := echo.Marshal(nil)
		if err != nil {
			errorHandler("Failed to marshal ICMP echo message", err.Error())
		}
		if _, err := conn.WriteTo(marEcho, dst); err != nil {
			errorHandler("Failed to send ICMP echo request", err.Error())
		}
		reply := make([]byte, 1500)
		if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			errorHandler("Failed to set read deadline", err.Error())
		}
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			fmt.Printf("Request timed out for icmp_seq %d\n", seq)
			time.Sleep(time.Second)
			continue
		}
		duration := time.Since(start)
		response, err := icmp.ParseMessage(1, reply[:n])
		if err != nil {
			errorHandler("Failed to parse ICMP response", err.Error())
		}

		switch response.Type {
		case ipv4.ICMPTypeEchoReply:
			fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n", n, peer, seq, duration)
		default:
			fmt.Printf("Received unexpected ICMP message type: %v\n", response.Type)
		}
		time.Sleep(time.Second)
	}
}

func errorHandler(context, err string) {
	log.Printf("Error: %s - %s\n", context, err)
	os.Exit(1)
}
