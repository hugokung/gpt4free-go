package main

import (
	"G4f/g4f"
	"G4f/g4f/provider"
	"fmt"
	"time"
)

func main() {
	msg := provider.Messages{
		{"role": "assistant", "content": "Hi! How can I help you?", "who": "AI: ", "timestamp": 0, "id": ""},
		{"role": "user", "content": "are you gpt-4?", "who": "User: ", "timestamp": 0, "id": ""},
	}
	for i := range msg {
		msg[i]["id"] = g4f.GetRandomString(11)
		msg[i]["timestamp"] = time.Now().Unix()
	}
	chat := provider.Chatgpt4Online{
		BaseProvider: provider.BaseProvider{},
	}
	recvCh := make(chan string)
	errCh := make(chan error)
	go chat.CreateAsyncGenerator(msg, recvCh, errCh)
	for {
		select {
		case resp := <-recvCh:
			fmt.Printf("%s\n", resp)
			//return
		case err := <-errCh:
			fmt.Printf("%v\n", err)
			return
		}
	}
}
