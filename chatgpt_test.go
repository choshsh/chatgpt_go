package chatgpt_go

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompletion(t *testing.T) {
	result, err := Completion("report for google", &CompletionConfig{
		MaxTokens: Int(50),
	})
	assert.NoError(t, err)
	assert.Greater(t, len(result.Choices), 0)
	assert.NotEqual(t, result.Choices[0].Text, "")
}

func TestCompletionStream(t *testing.T) {
	err := CompletionStream("report for google", &CompletionConfig{
		MaxTokens: Int(50),
	})
	fmt.Println()
	assert.NoError(t, err)
}

func TestSuffixWhenStream(t *testing.T) {
	str := `data: {"id": "cmpl-6gBfRPRNIyt5s47Y1BH0W1B4lEWdj", "object": "text_completion", "created": 1675512645, "choices": [{"text": "", "index": 0, "logprobs": null, "finish_reason": "stop"}], "model": "text-davinci-003"}`

	var m ChatGptStream
	err := json.Unmarshal([]byte(str[5:]), &m)

	assert.NoError(t, err)
	assert.Greater(t, len(m.Choices), 0)
	assert.Equal(t, m.Choices[0].FinishReason, "stop")
}
