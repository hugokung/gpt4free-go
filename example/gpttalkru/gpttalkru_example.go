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

	chat := provider.GptTalkRu{
		BaseProvider: provider.BaseProvider{
			BaseUrl:  "https://gpttalk.ru",
			ProxyUrl: "127.0.0.1:7890",
		},
	}

	recvCh := make(chan string)
	errCh := make(chan error)

	go chat.CreateAsyncGenerator(msg, recvCh, errCh)
	for {
		select {
		case res := <-recvCh:
			fmt.Printf("res: %v", res)
			return
		case err := <-errCh:
			fmt.Printf("err: %v", err)
			return
		}
	}
}
