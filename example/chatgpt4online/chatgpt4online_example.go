package main

import (
	"fmt"
	"time"

	"github.com/hugokung/G4f/g4f/provider"
	"github.com/hugokung/G4f/g4f/utils"
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
		BaseProvider: &provider.BaseProvider{
			BaseUrl: "https://chatgpt4online.org",
		},
	}
	recvCh := make(chan string)
	errCh := make(chan error)
	go chat.CreateAsyncGenerator(msg, recvCh, errCh, "", true, nil)
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
