package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

type Chat struct {
	Username string

	incomingMessages chan Message
	outgoingMessages chan Message
	uiRefreshTrigger chan bool
	messages         []Message
	mutex            sync.Mutex
}

func clearScreen() {
    var clearCmd *exec.Cmd
    if runtime.GOOS == "windows" {
        clearCmd = exec.Command("cmd", "/c", "cls")
    } else {
        clearCmd = exec.Command("clear")
    }
    clearCmd.Stdout = os.Stdout
    clearCmd.Run()
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
		var incomingMessage Message
		json.Unmarshal([]byte(strings.TrimSpace(msg)), &incomingMessage)
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
		msg := Message{Sender: c.Username, Text: text, Timestamp: time.Now().Format(time.RFC3339)}
		c.outgoingMessages <- msg
		c.mutex.Lock()
		c.messages = append(c.messages, msg)
		c.mutex.Unlock()
		c.uiRefreshTrigger <- true // Trigger UI refresh
	}
}

func (c *Chat) Run(listenAddress, peerAddress string) {
	// Listen for incoming connections
	go func() {
		ln, err := net.Listen("tcp", listenAddress)
		if err != nil {
			fmt.Println("Error setting up listener:", err.Error())
			return
		}
		defer ln.Close()

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err.Error())
				continue
			}
			go c.handleIncoming(conn)
		}
	}()

	// Connect to the peer
	go func() {
		var conn net.Conn
		var err error
		for {
			conn, err = net.Dial("tcp", peerAddress)
			if err == nil {
				go c.handleOutgoing(conn)
				break
			}
			fmt.Println("Error connecting:", err.Error(), "; retrying in a second...")
			time.Sleep(time.Second)
			clearScreen()
		}
	}()

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

func main() {
	listenIP := flag.String("listen-ip", "localhost", "IP to listen on")
	listenPort := flag.String("listen-port", "8080", "Port to listen on")
	peerIP := flag.String("peer-ip", "localhost", "IP of the peer to connect to")
	peerPort := flag.String("peer-port", "8081", "Port of the peer to connect to")
	username := flag.String("username", "User", "Your username in the chat")
	flag.Parse()

	listenAddress := *listenIP + ":" + *listenPort
	peerAddress := *peerIP + ":" + *peerPort

	chat := &Chat{
		Username:         *username,
		incomingMessages: make(chan Message),
		outgoingMessages: make(chan Message),
		uiRefreshTrigger: make(chan bool),
	}

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
