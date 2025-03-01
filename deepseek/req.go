package deepseek

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type DSRequest struct {
	Model     string
	Messages  []map[string]string
	MaxTokens int
	Api       string
}

type DSResponse struct {
	ID                string     `json:"id"`
	Object            string     `json:"object"`
	Created           int        `json:"created"`
	Model             string     `json:"model"`
	SystemFingerPrint string     `json:"system_fingerprint"`
	Choices           []DSChoice `json:"choices"`
}

type DSChoice struct {
	Delta struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	}
}

func NewDSRequest(api string) *DSRequest {
	return &DSRequest{
		Model:     "deepseek-chat",
		Messages:  make([]map[string]string, 0),
		MaxTokens: 2048,
		Api:       api,
	}
}

func (r *DSRequest) ClearMsg() {
	r.Messages = make([]map[string]string, 0)
	clear(r.Messages)
}

func (r *DSRequest) AddUserMsg(content string) {
	r.Messages = append(r.Messages, map[string]string{
		"content": content,
		"role":    "user",
	})

}

func (r *DSRequest) AddAssistantMsg(content string) {
	r.Messages = append(r.Messages, map[string]string{
		"content": content,
		"role":    "assistant",
	})
}

func (r *DSRequest) SimulateSend(respWriter *io.PipeWriter) {

	fmt.Println(r.Messages)

	defer respWriter.Close()

	ch := make(chan string)
	msgs := strings.Split("This is a test message for simulate response from deepseek", " ")
	go func() {
		defer close(ch)
		for _, msg := range msgs {
			ch <- msg + " "
			time.Sleep(100 * time.Millisecond)
		}
	}()

	for data := range ch {
		_, err := respWriter.Write([]byte(data))
		if err != nil {
			log.Fatal("error writing resp", err)
		}
	}

}

func (r *DSRequest) Send(respWriter *io.PipeWriter) {

	url := "https://api.deepseek.com/chat/completions"
	method := "POST"
	assistantContent := ""

	jsonData, err := json.Marshal(r.Messages)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	payload := strings.NewReader(fmt.Sprintf(`{
  "messages": %v,
  "model": "deepseek-chat",
  "frequency_penalty": 0,
  "max_tokens": %v,
  "presence_penalty": 0,
  "response_format": {
    "type": "text"
  },
  "stop": null,
  "stream": true,
  "stream_options": {
	  "include_usage": false
  },
  "temperature": 1,
  "top_p": 1,
  "tools": null,
  "tool_choice": "none",
  "logprobs": false,
  "top_logprobs": null
}`, string(jsonData), r.MaxTokens))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.Api))

	res, err := client.Do(req)
	if err != nil {
		log.Println("send req error:", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", res.StatusCode)
	}

	reader := bufio.NewReader(res.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				r.AddAssistantMsg(assistantContent)
				respWriter.Close()
				break
			}
			log.Fatalf("Error reading response: %v", err)
		}

		line = bytes.TrimSpace(line)

		// empty line is reached, meaining end of an event
		if len(line) == 0 {
			continue
		}

		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		data := bytes.TrimSpace(line[5:])
		if bytes.Equal(data, []byte("[DONE]")) {
			continue
		}
		var resp DSResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Println("json unmarshal error", err)
			continue
		}
		content := resp.Choices[0].Delta.Content
		// role := resp.Choices[0].Delta.Role
		assistantContent += content
		_, err = respWriter.Write([]byte(content))
		if err != nil {
			log.Fatal("Error writing to respWriter", err)
		}

	}
}
