package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// Channel for incoming messages
var incomingMessages = make(chan string)

// Channel for outgoing messages
var outgoingMessages = make(chan string)

// Function for handling incoming connections
func handleConnection(conn net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		msg, _ := reader.ReadString('\n')
		incomingMessages <- msg
	}
}

// Function for handling outgoing connections
func handleOutgoingConnection(conn net.Conn) {
	for {
		msg := <-outgoingMessages
		fmt.Fprintf(conn, msg+"\n")
	}
}

// Function for reading user input
func userInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message: ")
		text, _ := reader.ReadString('\n')
		outgoingMessages <- text
	}
}

func main() {
	// Accept incoming connections
	ln, _ := net.Listen("tcp", ":8080")
	go func() {
		for {
			conn, _ := ln.Accept()
			go handleConnection(conn)
			go handleOutgoingConnection(conn)
		}
	}()

	// Connect to the peer
	conn, _ := net.Dial("tcp", "localhost:8081")
	go handleConnection(conn)
	go handleOutgoingConnection(conn)

	// Handle user input
	go userInput()

	// Print incoming messages
	for {
		msg := <-incomingMessages
		fmt.Print("Received message: ", msg)
	}
}
