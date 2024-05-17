package main

import (
	"G4f/g4f/provider"
	"G4f/g4f/utils"
	"fmt"
	"time"
)

// xvfb-run go run chatgpt4online_example.go
func main() {
	msg := provider.Messages{
		{"role": "assistant", "content": "Hi! How can I help you?", "who": "AI: ", "timestamp": 0, "id": ""},
		{"role": "user", "content": "are you gpt-4?", "who": "User: ", "timestamp": 0, "id": ""},
	}
	for i := range msg {
		msg[i]["id"] = utils.GetRandomString(11)
		msg[i]["timestamp"] = time.Now().Unix()
	}
	chat := provider.Chatgpt4Online{
		BaseProvider: provider.BaseProvider{
			BaseUrl:  "https://chatgpt4online.org",
			ProxyUrl: "",
		},
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
