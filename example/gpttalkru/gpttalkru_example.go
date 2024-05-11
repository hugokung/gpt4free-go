package main

import (
	"G4f/g4f/provider"
	"fmt"
)

func main() {

	msg := provider.Messages{
		{"role": "assistant", "content": "Hi! How can I help you?"},
		{"role": "user", "content": "are you gpt-4?"},
	}

	chat := provider.GptTalkRu{
		BaseProvider: provider.BaseProvider{
			BaseUrl:  "https://gpttalk.ru",
			ProxyUrl: "",
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
