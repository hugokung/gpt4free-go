package g4f

type Message []map[string]string

type ChatRequest struct {
	Messages          Message `json:"messages"`
	Temperature       float32 `json:"temperature,omitempty"`
	Top_p             int     `json:"top_p,omitempty"`
	N                 int     `json:"n,omitempty"`
	Stream            bool    `json:"stream,omitempty"`
	Max_tokens        int     `json:"max_tokens,omitempty"`
	Presence_penalty  int     `json:"presence_penalty,omitempty"`
	Frequency_penalty int     `json:"frequency_penalty,omitempty"`
	User              string  `json:"user"`
}
