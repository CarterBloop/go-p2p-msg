package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"sync"
	"encoding/json"
	"strings"
	"time"
)

type Chat struct {
	Username string

	incomingMessages chan Message
	outgoingMessages chan Message
	uiRefreshTrigger chan bool
	messages         []Message
	mutex            sync.Mutex
}

func NewChat(username string) *Chat {
	return &Chat{
		Username:         username,
		incomingMessages: make(chan Message),
		outgoingMessages: make(chan Message),
		uiRefreshTrigger: make(chan bool),
	}
}

func (c *Chat) handleIncoming(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err.Error())
			conn.Close()
			return
		}
		incomingMessage, _ := parseMessage(msg)
		c.mutex.Lock()
		c.incomingMessages <- incomingMessage
		c.mutex.Unlock()
		c.uiRefreshTrigger <- true // Trigger UI refresh
	}

}

func (c *Chat) handleOutgoing(conn net.Conn) {
	for {
		msg := <-c.outgoingMessages
		msgJSON, _ := json.Marshal(msg)
		fmt.Fprintf(conn, string(msgJSON)+"\n")
	}
}

func (c *Chat) userInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from user:", err.Error())
			return
		}
		text = strings.TrimSpace(text)
		msg := createMessage(c.Username, text)
		c.outgoingMessages <- msg
		c.mutex.Lock()
		c.messages = append(c.messages, msg)
		c.mutex.Unlock()
		c.uiRefreshTrigger <- true // Trigger UI refresh
	}
}

func (c *Chat) Run(listenAddress, peerAddress string) {
	go listenForConnections(listenAddress, c)
	go connectToPeer(peerAddress, c)

	// Handle user input
	go c.userInput()

	// Colors for incoming and outgoing messages
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for {
		select {
		case msg := <-c.incomingMessages:
			c.mutex.Lock()
			c.messages = append(c.messages, msg)
			c.mutex.Unlock()
		case <-c.uiRefreshTrigger:
			// Clear the console
			clearScreen()

			// Print the conversation
			for _, msg := range c.messages {
				if msg.Sender == c.Username {
					fmt.Println(green("You:"), msg.Text)
				} else {
					fmt.Println(blue(msg.Sender+":"), msg.Text)
				}
			}
		}
	}
}

func listenForConnections(listenAddress string, chat *Chat) {
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		fmt.Println("Error setting up listener:", err.Error())
		return
	}
	defer ln.Close()

	fmt.Printf("Listening for connections on %s...\n", listenAddress)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())
		go chat.handleIncoming(conn)
	}
}

func connectToPeer(peerAddress string, chat *Chat) {
	var conn net.Conn
	var err error
	for {
		fmt.Printf("Attempting to connect to peer at %s...\n", peerAddress)
		conn, err = net.Dial("tcp", peerAddress)
		if err == nil {
			fmt.Printf("Successfully connected to peer at %s\n", peerAddress)
			go chat.handleOutgoing(conn)
			break
		}
		fmt.Println("Error connecting:", err.Error(), "; retrying in a second...")
		time.Sleep(time.Second)
		clearScreen()
	}
}
