package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL      string
	apiKey       string
	SystemPrompt string
	Model        string
	httpClient   *http.Client
}

func NewClient(baseURL, apiKey, systemPrompt, model string) *Client {
	return &Client{
		baseURL:      baseURL,
		apiKey:       apiKey,
		SystemPrompt: systemPrompt,
		Model:        model,
		httpClient:   http.DefaultClient,
	}
}

func (c *Client) CreateChatCompletion(name string, maxCharacters int) ([]byte, error) {
	messages := []map[string]string{
		{
			"role":    "system",
			"content": fmt.Sprintf(
				"%s\n\n" +
					"ВНИМАНИЕ! Ответ ДОЛЖЕН быть короче %d символов (включая пробелы и знаки препинания). " +
					"Если превысишь лимит — ответ будет отключен.",
				c.SystemPrompt,
				maxCharacters),
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Название: %s", name),
		},
	}
	requestBody := map[string]interface{}{
		"model":    c.Model,
		"messages": messages,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
