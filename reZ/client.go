// Package reZ provides a client for interacting with Zhipu AI's GLM-4.6 model
// through reverse-engineered API endpoints.
//
// Features:
//   - Streaming responses (SSE)
//   - Conversation history management
//   - Thinking mode (shows AI reasoning)
//   - Web search integration
//   - Tool/Function calling
//   - System prompts
//   - Flexible configuration (temperature, max_tokens, top_p)
//
// Example usage:
//
//	client := reZ.NewClient()
//	ctx := context.Background()
//
//	events, err := client.Chat(ctx, "Hello!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for event := range events {
//	    if event.Error != nil {
//	        log.Fatal(event.Error)
//	    }
//	    fmt.Print(event.Text)
//	}
package reZ

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	baseURL = "https://www.bigmodel.cn/api/biz/trial/response/v4/sse/11170"
	
	// Hardcoded credentials
	authToken    = "eyJhbGciOiJIUzUxMiJ9.eyJ1c2VyX3R5cGUiOiJQRVJTT05BTCIsInVzZXJfaWQiOjMyNzM1NjAsInVzZXJfa2V5IjoiYTc3YWExOGMtNWU3OS00MmY2LWEzMzUtMDdkOThkNzAxZGRiIiwiY3VzdG9tZXJfaWQiOiI4MTY1MTc1OTUzNzc3MTE2OCIsInVzZXJuYW1lIjoid3lkZ2p0NDgifQ.qk-siKEmVRuHfeBVmgmjzD4yqkM48QIzPXMRMzu3dcAHAgexbeAfNwTUIeRe0yHQqfr79nticLWAS7DWrHDIEQ"
	organization = "org-f2ADE5b25C3B4cF789a3C75A0ec80357"
	project      = "proj_dfAE017A77744573bAA86200824B45e0"
	cookieValue  = "sensorsdata2015jssdkchannel=%7B%22prop%22%3A%7B%22_sa_channel_landing_url%22%3A%22%22%7D%7D; sensorsdata2015jssdkcross=%7B%22distinct_id%22%3A%2281651759537771168%22%2C%22first_id%22%3A%22199792d9dd7230-0a1fe2c8720f078-43330223-2073600-199792d9dd81ba%22%2C%22props%22%3A%7B%22%24latest_traffic_source_type%22%3A%22%E7%9B%B4%E6%8E%A5%E6%B5%81%E9%87%8F%22%2C%22%24latest_search_keyword%22%3A%22%E6%9C%AA%E5%8F%96%E5%88%B0%E5%80%BC_%E7%9B%B4%E6%8E%A5%E6%89%93%E5%BC%80%22%2C%22%24latest_referrer%22%3A%22%22%2C%22%24latest_utm_source%22%3A%22bigModel%22%2C%22%24latest_utm_medium%22%3A%22Experience-Center%22%2C%22%24latest_utm_campaign%22%3A%22Platform_Ops%22%2C%22%24latest_utm_content%22%3A%22glm-code%22%7D%2C%22identities%22%3A%22eyIkaWRlbnRpdHlfY29va2llX2lkIjoiMTk5NzkyZDlkZDcyMzAtMGExZmUyYzg3MjBmMDc4LTQzMzMwMjIzLTIwNzM2MDAtMTk5NzkyZDlkZDgxYmEiLCIkaWRlbnRpdHlfbG9naW5faWQiOiI4MTY1MTc1OTUzNzc3MTE2OCJ9%22%2C%22history_login_id%22%3A%7B%22name%22%3A%22%24identity_login_id%22%2C%22value%22%3A%2281651759537771168%22%7D%7D; sensorsdata2015jssdksession=%7B%22session_id%22%3A%22199ac9dba9264b0e34a8475b5f968433302232073600199ac9dba93787%22%2C%22first_session_time%22%3A1759537642129%2C%22latest_session_time%22%3A1759538303517%7D; acw_tc=ac11000117595376309637503edbce204d13e2d1567e380e4943fb80ea424f; bigmodel_token_production=eyJhbGciOiJIUzUxMiJ9.eyJ1c2VyX3R5cGUiOiJQRVJTT05BTCIsInVzZXJfaWQiOjMyNzM1NjAsInVzZXJfa2V5IjoiYTc3YWExOGMtNWU3OS00MmY2LWEzMzUtMDdkOThkNzAxZGRiIiwiY3VzdG9tZXJfaWQiOiI4MTY1MTc1OTUzNzc3MTE2OCIsInVzZXJuYW1lIjoid3lkZ2p0NDgifQ.qk-siKEmVRuHfeBVmgmjzD4yqkM48QIzPXMRMzu3dcAHAgexbeAfNwTUIeRe0yHQqfr79nticLWAS7DWrHDIEQ; ph_phc_TXdpocbGVeZVm5VJmAsHTMrCofBQu3e0kN8HGMNGTVW_posthog=%7B%22distinct_id%22%3A%2201997935-9e8d-7b7b-9aeb-5f7fa4e9f512%22%2C%22%24sesid%22%3A%5B1759537971666%2C%220199aca2-c14c-7350-ac9b-c1b6bfa0867f%22%2C1759537971532%5D%7D"
	
	// Configuration constants
	defaultChannelBuffer = 100
	defaultMaxTokens     = 65536
	defaultTemperature   = 1.0
	defaultTopP          = 0.95
)

// Client represents a client for interacting with Zhipu AI API.
// It manages HTTP connections, conversation history, and streaming responses.
// The client is thread-safe and can be used concurrently from multiple goroutines.
type Client struct {
	httpClient *http.Client
	history    []Message
	mu         sync.RWMutex // Protects history
}

// NewClient creates a new Zhipu AI client instance.
// The client is ready to use immediately with default settings.
//
// Example:
//
//	client := reZ.NewClient()
//	events, err := client.Chat(context.Background(), "Hello!")
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		history:    make([]Message, 0),
	}
}

// Chat sends a message and returns a channel of streaming events.
// This method does not maintain conversation history.
// Use ChatWithHistory for conversations that need context.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - content: The message to send
//   - opts: Optional configuration (WithSystemPrompt, WithTemperature, etc.)
//
// Returns:
//   - Channel of StreamEvent objects
//   - Error if the request fails
//
// Example:
//
//	events, err := client.Chat(ctx, "Hello!",
//	    reZ.WithTemperature(0.7),
//	    reZ.WithMaxTokens(1000))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for event := range events {
//	    fmt.Print(event.Text)
//	}
func (c *Client) Chat(ctx context.Context, content string, opts ...Option) (<-chan StreamEvent, error) {
	req := c.buildRequest(content, opts...)
	return c.stream(ctx, req)
}

// ChatWithHistory sends a message with automatic conversation history management.
// The client maintains the conversation context across multiple calls.
// Pass an empty string for content when continuing after a tool call response.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - content: The message to send (can be empty for tool response continuation)
//   - opts: Optional configuration
//
// Returns:
//   - Channel of StreamEvent objects
//   - Error if the request fails
//
// Example:
//
//	// First message
//	events, _ := client.ChatWithHistory(ctx, "My name is Alex")
//	for event := range events {
//	    fmt.Print(event.Text)
//	}
//
//	// Second message - AI remembers context
//	events, _ = client.ChatWithHistory(ctx, "What's my name?")
//	for event := range events {
//	    fmt.Print(event.Text) // "Your name is Alex"
//	}
func (c *Client) ChatWithHistory(ctx context.Context, content string, opts ...Option) (<-chan StreamEvent, error) {
	c.mu.Lock()
	if content != "" {
		c.history = append(c.history, Message{
			Role:            "user",
			Content:         content,
			FileContentList: []interface{}{},
		})
	}
	
	req := c.buildRequest(content, opts...)
	req.Prompt = c.history
	c.mu.Unlock()
	
	eventCh := make(chan StreamEvent, defaultChannelBuffer)
	
	go func() {
		defer close(eventCh)
		
		respCh, err := c.stream(ctx, req)
		if err != nil {
			eventCh <- StreamEvent{Error: err}
			return
		}
		
		var fullResponse strings.Builder
		var toolCalls []ToolCall
		
		for event := range respCh {
			eventCh <- event
			
			if event.Error != nil {
				return
			}
			
			if event.Text != "" {
				fullResponse.WriteString(event.Text)
			}
			
			if event.ToolCall != nil {
				toolCalls = append(toolCalls, *event.ToolCall)
			}
		}
		
		// Add assistant response to history
		if fullResponse.Len() > 0 || len(toolCalls) > 0 {
			msg := Message{Role: "assistant"}
			if fullResponse.Len() > 0 {
				msg.Content = fullResponse.String()
			}
			if len(toolCalls) > 0 {
				msg.ToolCalls = toolCalls
			}
			c.mu.Lock()
			c.history = append(c.history, msg)
			c.mu.Unlock()
		}
	}()
	
	return eventCh, nil
}

// AddToolResponse adds a tool/function response to the conversation history.
// Use this after the AI requests a function call via ToolCall event.
//
// Parameters:
//   - content: The result from the function execution (typically JSON)
//
// Example:
//
//	for event := range events {
//	    if event.ToolCall != nil {
//	        result := executeFunction(event.ToolCall.Function.Name)
//	        client.AddToolResponse(result)
//	        // Continue conversation
//	        client.ChatWithHistory(ctx, "")
//	    }
//	}
func (c *Client) AddToolResponse(content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = append(c.history, Message{
		Role:    "tool",
		Content: content,
	})
}

// ClearHistory clears the conversation history.
// Use this to start a fresh conversation without previous context.
//
// Example:
//
//	client.ClearHistory()
//	client.ChatWithHistory(ctx, "New topic")
func (c *Client) ClearHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = make([]Message, 0)
}

// GetHistory returns the current conversation history.
// The history includes all user messages, assistant responses, and tool calls.
//
// Returns:
//   - Slice of Message objects representing the conversation
//
// Example:
//
//	history := client.GetHistory()
//	for i, msg := range history {
//	    fmt.Printf("%d. [%s]: %s\n", i+1, msg.Role, msg.Content)
//	}
func (c *Client) GetHistory() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Return a copy to prevent external modifications
	historyCopy := make([]Message, len(c.history))
	copy(historyCopy, c.history)
	return historyCopy
}

func (c *Client) buildRequest(content string, opts ...Option) *Request {
	req := &Request{
		Model:   "glm-4.6",
		ModelID: 11170,
		Prompt: []Message{
			{
				Role:            "user",
				Content:         content,
				FileContentList: []interface{}{},
			},
		},
		Stream:      true,
		Thinking:    &Thinking{Type: "enabled"},
		MaxTokens:   defaultMaxTokens,
		Temperature: defaultTemperature,
		TopP:        defaultTopP,
	}
	
	for _, opt := range opts {
		opt(req)
	}
	
	return req
}

func (c *Client) stream(ctx context.Context, req *Request) (<-chan StreamEvent, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	c.setHeaders(httpReq)
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	eventCh := make(chan StreamEvent, 100)
	
	go func() {
		defer close(eventCh)
		defer resp.Body.Close()
		
		scanner := bufio.NewScanner(resp.Body)
		var currentEvent string
		var currentData string
		
		for scanner.Scan() {
			line := scanner.Text()
			
			if line == "" {
				if currentEvent != "" && currentData != "" {
					c.processEvent(currentEvent, currentData, eventCh)
					currentEvent = ""
					currentData = ""
				}
				continue
			}
			
			if strings.HasPrefix(line, "event:") {
				currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			} else if strings.HasPrefix(line, "data:") {
				currentData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			}
		}
		
		if err := scanner.Err(); err != nil && err != io.EOF {
			eventCh <- StreamEvent{Error: fmt.Errorf("scan error: %w", err)}
		}
	}()
	
	return eventCh, nil
}

func (c *Client) processEvent(event, data string, ch chan<- StreamEvent) {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		ch <- StreamEvent{Error: fmt.Errorf("unmarshal event data: %w", err)}
		return
	}
	
	streamEvent := StreamEvent{
		Event: event,
		Raw:   payload,
	}
	
	if think, ok := payload["think"].(string); ok {
		streamEvent.Think = think
	}
	
	if text, ok := payload["text"].(string); ok {
		streamEvent.Text = text
	}
	
	// Tool call parsing (event: functionHit)
	if tcMap, ok := payload["tool_calls"].(map[string]interface{}); ok {
		toolCall := &ToolCall{}
		
		if id, ok := tcMap["id"].(string); ok {
			toolCall.ID = id
		}
		if tcType, ok := tcMap["type"].(string); ok {
			toolCall.Type = tcType
		}
		if index, ok := tcMap["index"].(float64); ok {
			toolCall.Index = int(index)
		}
		
		if fn, ok := tcMap["function"].(map[string]interface{}); ok {
			toolCall.Function = &Function{
				Name:      fn["name"].(string),
				Arguments: fn["arguments"].(string),
			}
		}
		
		streamEvent.ToolCall = toolCall
	}
	
	ch <- streamEvent
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:142.0) Gecko/20100101 Firefox/142.0")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Set("Referer", "https://www.bigmodel.cn/trialcenter/modeltrial/text")
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Bigmodel-Organization", organization)
	req.Header.Set("Bigmodel-Project", project)
	req.Header.Set("Set-Language", "en")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.bigmodel.cn")
	req.Header.Set("Cookie", cookieValue)
}
