package chatgpt_go

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/joho/godotenv"
	"io"
	"log"
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
	defer resp.Body.Close()

	streamMessage, done, errorChannel := make(chan string), make(chan struct{}), make(chan error)
	go handleCompletionStream(resp, streamMessage, done, errorChannel)

	for {
		select {
		case msg := <-streamMessage:
			fmt.Print(msg)
		case <-done:
			return nil
		case e := <-errorChannel:
			return e
		case <-time.After(1 * time.Minute):
			return errors.New("context deadline exceeded")
		}
	}
}

func handleCompletionStream(resp *http.Response, streamMessage chan string, done chan struct{}, errorChannel chan error) {
	defer resp.Body.Close()
	defer close(streamMessage)
	defer close(done)
	defer close(errorChannel)

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		line = bytes.TrimSuffix(line, []byte(`\n`))
		if isEofText(line) {
			done <- struct{}{}
			return
		}

		// TrimPrefix "data: "
		var chatGptStream ChatGptStream
		err = json.Unmarshal(bytes.TrimPrefix(line, []byte(`data: `)), &chatGptStream)
		if err != nil {
			errorChannel <- err
			return
		}

		if isStreamStopped(chatGptStream) {
			done <- struct{}{}
			return
		}

		streamMessage <- chatGptStream.Choices[0].Text
	}
}

func isEofText(line []byte) bool {
	return bytes.Equal(line, []byte(EofText))
}

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
	if resp.StatusCode != http.StatusOK {
		rawBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(rawBytes))
	}

	return resp, nil
}
