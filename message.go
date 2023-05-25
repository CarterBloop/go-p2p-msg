package main

import (
	"encoding/json"
	"strings"
	"time"
)

type Message struct {
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

func parseMessage(data string) (Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(strings.TrimSpace(data)), &msg)
	return msg, err
}

func createMessage(sender, text string) Message {
	return Message{
		Sender:    sender,
		Text:      text,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
