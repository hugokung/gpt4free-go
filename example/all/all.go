package main

import (
	"G4f/g4f"
	"G4f/g4f/client"
	"G4f/g4f/provider"
	"errors"
	"log"
)

func main() {
	msg := provider.Messages{
		{"role": "user", "content": "hello, can you help me?"},
	}
	cli, err := client.Client{}.Create(msg, "gpt-3.5-turbo", "", true, "", 60, "", false, false)
	if err != nil {
		return
	}

	for {
		select {
		case err := <-cli.ErrCh:
			log.Printf("cli.ErrCh: %v\n", err)
			if errors.Is(err, g4f.ErrStreamRestart) {
				continue
			}
			return
		case data := <-cli.RecCh:
			log.Printf("recv data: %v\n", data)
		}
	}
}
