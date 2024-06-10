package main

import (
	"fmt"

	"github.com/hugokung/G4f/g4f/provider"
)

func main() {

	msg := provider.Messages{
		{"role": "system", "content": "You are a helpful assistant."},
		{"role": "user", "content": "write a golang program"},
	}

	var llama *provider.Llama
	chat := llama.Create()

	RecvCh := make(chan string)
	ErrCh := make(chan error)

	go chat.CreateAsyncGenerator(msg, RecvCh, ErrCh, "http://127.0.0.1:7890", true, nil)
	for {
		select {
		case res := <-RecvCh:
			fmt.Printf("res: %v", res)
			//return
		case err := <-ErrCh:
			fmt.Printf("err: %v\n", err)
			return
		}
	}
}
