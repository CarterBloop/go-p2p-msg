package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
    "os/exec"
    "runtime"
	"strings"
	"time"
)

type Message struct {
	Text     string
	Outgoing bool
}

var incomingMessages = make(chan Message)
var outgoingMessages = make(chan Message)
var uiRefreshTrigger = make(chan bool)
var messages = []Message{}

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

// handleIncoming modified with uiRefreshTrigger
func handleIncoming(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err.Error())
			return
		}
		incomingMessages <- Message{strings.TrimSpace(msg), false}
		uiRefreshTrigger <- true // Trigger UI refresh
	}
}

func handleOutgoing(conn net.Conn) {
	for {
		msg := <-outgoingMessages
		fmt.Fprintf(conn, msg.Text+"\n")
	}
}

func userInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from user:", err.Error())
			return
		}
		text = strings.TrimSpace(text)
		msg := Message{text, true}
		outgoingMessages <- msg
		messages = append(messages, msg)
		uiRefreshTrigger <- true // Trigger UI refresh
	}
}

func main() {
	listenAddress := flag.String("listen", "localhost:8080", "Address to listen on")
	peerAddress := flag.String("peer", "localhost:8081", "Address of the peer to connect to")
	flag.Parse()

	go func() {
		ln, err := net.Listen("tcp", *listenAddress)
		if err != nil {
			fmt.Println("Error setting up listener:", err.Error())
			return
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err.Error())
				continue
			}
			go handleIncoming(conn)
		}
	}()

	// Connect to the peer
	go func() {
		var conn net.Conn
		var err error
		for {
			conn, err = net.Dial("tcp", *peerAddress)
			if err == nil {
				go handleOutgoing(conn)
				break
			}
			fmt.Println("Error connecting:", err.Error(), "; retrying in a second...")
			time.Sleep(time.Second)
			clearScreen()
		}
	}()

	// Handle user input
	go userInput()

	// Colors for incoming and outgoing messages
	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for {
		select {
		case msg := <-incomingMessages:
			messages = append(messages, msg)
		case <-uiRefreshTrigger:
			// Clear the console
			clearScreen()

			// Print the conversation
			for _, msg := range messages {
				if msg.Outgoing {
					fmt.Println(green("You:"), msg.Text)
				} else {
					fmt.Println(blue("Friend:"), msg.Text)
				}
			}
		}
	}
}