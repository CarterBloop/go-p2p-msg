package main

import (
	"flag"
	"os"
	"os/signal"
	"fmt"
)

func main() {
	listenIP := flag.String("listen-ip", "localhost", "IP to listen on")
	listenPort := flag.String("listen-port", "8080", "Port to listen on")
	peerIP := flag.String("peer-ip", "localhost", "IP of the peer to connect to")
	peerPort := flag.String("peer-port", "8081", "Port of the peer to connect to")
	username := flag.String("username", "User", "Your username in the chat")
	flag.Parse()

	listenAddress := *listenIP + ":" + *listenPort
	peerAddress := *peerIP + ":" + *peerPort

	chat := NewChat(*username)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			fmt.Println("\r- Ctrl+C pressed in Terminal")
			os.Exit(0)
		}
	}()

	chat.Run(listenAddress, peerAddress)
}
