package main

import (
	"fmt"

	"github.com/hugokung/gpt4free-go/g4f/provider"
)

// xvfb-run go run gpttalkru_example.go
func main() {

	msg := provider.Messages{
		{"role": "assistant", "content": "Hi! How can I help you?"},
		{"role": "user", "content": "are you gpt-4?"},
	}

	chat := provider.GptTalkRu{
		BaseProvider: &provider.BaseProvider{
			BaseUrl: "https://gpttalk.ru",
		},
	}

	recvCh := make(chan string)
	errCh := make(chan error)

	go chat.CreateAsyncGenerator(msg, recvCh, errCh, "", true, nil)
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
