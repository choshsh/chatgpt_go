package chatgpt_go

import (
	"errors"
	"strings"
)

var (
	DefaultModel       = "text-davinci-003"
	DefaultToken       = 1000
	DefaultTemperature = 1.0
)

// ModelsResponse ref: https://platform.openai.com/docs/api-reference/models/list
type ModelsResponse struct {
	Data []struct {
		Id         string        `json:"id"`
		Object     string        `json:"object"`
		OwnedBy    string        `json:"owned_by"`
		Permission []interface{} `json:"permission"`
	} `json:"data"`
	Object string `json:"object"`
}

// completionPayload ref: https://platform.openai.com/docs/api-reference/completions/create
type completionPayload struct {
	Model            *string  `json:"model" binding:"required,oneof=text-davinci-003 text-curie-001 text-babbage-001 text-ada-001" default:"text-davinci-003"`
	Prompt           *string  `json:"prompt" binding:"required"`
	MaxTokens        *int     `json:"max_tokens" default:"16,min=-1,max=2048"`
	Temperature      *float64 `json:"temperature" default:"1.0"`
	TopP             *float64 `json:"top_p,omitempty"  default:"1.0"`
	N                *float64 `json:"n,omitempty" default:"1.0"`
	Stream           *bool    `json:"stream,omitempty"`
	Echo             *bool    `json:"echo,omitempty" default:"false"`
	Logprobs         *int     `json:"logprobs,omitempty" default:"0,max=5"`
	Stop             *string  `json:"stop,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty" default:"0,min=-2.0,max=2.0"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty" default:"0,min=-2.0,max=2.0"`
	BestOf           *int     `json:"best_of,omitempty,omitempty" default:"1"`
	User             *string  `json:"user,omitempty"`
}

// NewCompletionPayload Make CompletionPayload with default values.
func NewCompletionPayload(prompt string, config *CompletionConfig) (*completionPayload, error) {
	if p := strings.TrimSpace(prompt); len(p) == 0 {
		return nil, errors.New("prompt cannot be empty")
	}

	payload := &completionPayload{
		Model:       &DefaultModel,
		Prompt:      &prompt,
		MaxTokens:   &DefaultToken,
		Temperature: &DefaultTemperature,
	}

	if config != nil {
		if config.Model != nil {
			payload.Model = config.Model
		}
		if config.MaxTokens != nil {
			payload.MaxTokens = config.MaxTokens
		}
		if config.Temperature != nil {
			payload.Temperature = config.Temperature
		}
	}
	return payload, nil
}

// CompletionConfig CompletionPayload config
type CompletionConfig struct {
	Model       *string
	MaxTokens   *int
	Temperature *float64
}

// CompletionResponse Response from OpenAI API
type CompletionResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatGptStream Response from OpenAI API when using stream
// ref: https://platform.openai.com/docs/api-reference/completions/create#completions/create-stream
type ChatGptStream struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
}

type BaseErrorResponse struct {
	message *string
}
