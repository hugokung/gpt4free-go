package main

import (
	"G4f/g4f/provider"
	"fmt"
)

func main() {
	msg := provider.Messages{
		{"role": "system", "content": "You are a helpful assistant."},
		{"role": "user", "content": "are you gpt-4?"},
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
		case err := <-errCh:
			fmt.Printf("%v\n", err)
			return
		}
	}
}
