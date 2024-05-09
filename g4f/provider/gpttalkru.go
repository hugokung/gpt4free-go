package provider

import (
	"G4f/g4f"
	"log"
)

type GptTalkRu struct {
	BaseProvider
}

func (g *GptTalkRu) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error) {

	cookies, err := g4f.GetArgsFromBrowser(g.BaseUrl, g.ProxyUrl, 5, true)
	log.Printf("cookies: %v, err: %v\n", cookies, err)
	if err != nil {
		errCh <- err
		return
	}
	recvCh <- "test finished"
}
