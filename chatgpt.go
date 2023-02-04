package chatgpt_go

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"time"
)

var OpenAiToken string

func init() {
	godotenv.Load()
	var ok bool

	OpenAiToken, ok = os.LookupEnv("OPENAI_TOKEN")
	if !ok {
		panic("Cannot find OpenAi API Token [OPENAI_TOKEN]")
	}
}

// Completion ChatGPT-3 Text Generation
func Completion(prompt string, cfg *CompletionConfig) (*CompletionResponse, error) {
	payload, err := NewCompletionPayload(prompt, cfg)
	if err != nil {
		return nil, err
	}

	resp, err := SendMessage(payload)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	completionResponse := CompletionResponse{}
	err = json.Unmarshal(bytes, &completionResponse)
	if err != nil {
		return nil, err
	}

	return &completionResponse, nil
}

// CompletionStream ChatGPT-3 Text Generation with stream
func CompletionStream(prompt string, cfg *CompletionConfig) error {
	payload, err := NewCompletionPayload(prompt, cfg)
	if err != nil {
		return err
	}

	payload.Stream = Bool(true)
	resp, err := SendMessage(payload)
	if err != nil {
		return err
	}

	streamCh := make(chan string)
	errCh := make(chan error)
	done := make(chan struct{})

	go func() {
		defer resp.Body.Close()
		defer close(streamCh)
		defer close(errCh)
		defer close(done)

		for {
			line, errInCh := bufio.NewReader(resp.Body).ReadBytes('\n')
			if errInCh != nil {
				if errInCh.Error() == "EOF" {
					done <- struct{}{}
					break
				}
				errCh <- errInCh
				break
			}

			// suffix '\n'
			line = line[:len(line)-1]
			if bytes.Equal(line, []byte(EofText)) {
				done <- struct{}{}
				break
			}

			var chatGptStream ChatGptStream
			errInCh = json.Unmarshal(line[5:], &chatGptStream)
			if errInCh != nil {
				errCh <- errInCh
				break
			}

		if isStreamStopped(chatGptStream) {
				done <- struct{}{}
				break
			}

			streamCh <- chatGptStream.Choices[0].Text
		}
	}()

	for {
		// Print response
		select {
		case msg := <-streamCh:
			fmt.Print(msg)
		case <-done:
			return nil
		case e := <-errCh:
			return e
		case <-time.After(1 * time.Minute):
			return errors.New("context deadline exceeded")
func isStreamStopped(chatGptStream ChatGptStream) bool {
	stopReasons := []string{"stop", "length"}
	finishReason := chatGptStream.Choices[0].FinishReason

	for _, stopReason := range stopReasons {
		if finishReason == stopReason {
			fmt.Println()
			log.Printf("Completion is stopped due to [%s]\n", stopReason)
			return true
		}
	}
	return false
}

// SendMessage Request to OpenAI API and return *http.Response
func SendMessage(payload *completionPayload) (*http.Response, error) {
	var err error

	param, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	reqp, err := http.NewRequest(http.MethodPost, EndpointCompletion, bytes.NewBuffer(param))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: OpenAiTimeout * time.Second}

	reqp.Header.Add("Content-Type", "application/json")
	reqp.Header.Add("Authorization", OpenAiToken)

	resp, err := client.Do(reqp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
