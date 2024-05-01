package g4f

type ChatMessage []map[string]string

type ChatRequest struct {
	Messages         ChatMessage         `json:"messages"`
	Temperature      float32             `json:"temperature,omitempty"`
	TopP             int                 `json:"top_p,omitempty"`
	N                int                 `json:"n,omitempty"`
	Stream           bool                `json:"stream,omitempty"`
	Max_tokens       int                 `json:"max_tokens,omitempty"`
	PresencePenalty  float32             `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32             `json:"frequency_penalty,omitempty"`
	Model            string              `json:"model"`
	Seed             *int                `json:"seed,omitempty"`
	LogitBias        map[string]int      `json:"logit_bias,omitempty"`
	LogProbs         bool                `json:"logprobs,omitempty"`
	TopLogProbs      int                 `json:"top_logprobs,omitempty"`
	ResponseFormat   *ChatResponseFormat `json:"response_format,omitempty"`
}

type ChatResponseFormatType string

type ChatResponseFormat struct {
	Type ChatResponseFormatType `json:"type,omitempty"`
}

type ChatResponseChoices struct {
	Index        int                  `json:"index"`
	Message      *ChatResponseMessage `json:"message"`
	LogProbs     bool                 `json:"logprobs"`
	FinishReason string               `json:"finish_reason"`
}

type ChatResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type ChatResponse struct {
	ID                string                 `json:"id"`
	Choices           []*ChatResponseChoices `json:"choices"`
	Created           int                    `json:"created"`
	Model             string                 `json:"model"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	Object            string                 `json:"object"`
	Usage             *ChatResponseUsage     `json:"usage"`
}

type ChatStreamResponseFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatStreamResponseToolCalls struct {
	Index    int                         `json:"index"`
	ID       string                      `json:"id"`
	Type     string                      `json:"type"`
	Function *ChatStreamResponseFunction `json:"function"`
}

type ChatStreamResponseDelta struct {
	Content  string                       `json:"content"`
	ToolCall *ChatStreamResponseToolCalls `json:"tool_call"`
	Role     string                       `json:"role"`
}

type ChatStreamTopLogprobs struct {
	Token   string `json:"token"`
	Logprob int    `json:"logprob"`
	Bytes   []int  `json:"bytes"`
}

type ChatStreamLogprobsContent struct {
	Token       string                   `json:"token"`
	Logprob     int                      `json:"logprob"`
	Bytes       []int                    `json:"bytes"`
	TopLogprobs []*ChatStreamTopLogprobs `json:"top_logrobs"`
}

type ChatStreamResponseLogprobs struct {
	Content []*ChatStreamLogprobsContent `json:"content"`
}

type ChatStreamResponseChoices struct {
	Delta        *ChatStreamResponseDelta    `json:"delta"`
	Logprobs     *ChatStreamResponseLogprobs `json:"logprobs"`
	FinishReason string                      `json:"finish_reason"`
	Index        int                         `json:"index"`
}

type ChatStreamResponse struct {
	ID                string                       `json:"id"`
	Choices           []*ChatStreamResponseChoices `json:"choices"`
	Created           int                          `json:"created"`
	Model             string                       `json:"model"`
	SystemFingerprint string                       `json:"system_fingerprint"`
	Object            string                       `json:"object"`
}
