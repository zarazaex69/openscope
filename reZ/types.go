package reZ

import "fmt"

// Request represents an API request to Zhipu AI.
type Request struct {
	Model        string     `json:"model"`
	ModelID      int        `json:"modelId"`
	Prompt       []Message  `json:"prompt"`
	Stream       bool       `json:"stream"`
	Thinking     *Thinking  `json:"thinking,omitempty"`
	MaxTokens    int        `json:"max_tokens"`
	Temperature  float64    `json:"temperature"`
	TopP         float64    `json:"top_p"`
	SystemPrompt string     `json:"system_prompt,omitempty"`
	Tools        []Tool     `json:"tools,omitempty"`
}

// Message represents a message in the conversation.
// Role can be "user", "assistant", or "tool".
type Message struct {
	Role            string        `json:"role"`
	Content         string        `json:"content,omitempty"`
	FileContentList []interface{} `json:"fileContentList,omitempty"`
	ToolCalls       []ToolCall    `json:"toolCalls,omitempty"`
}

// Thinking configures the thinking mode settings.
type Thinking struct {
	Type string `json:"type"`
}

// Tool represents a tool (function or web_search) that the AI can use.
type Tool struct {
	Type      string     `json:"type"`
	Function  *Function  `json:"function,omitempty"`
	WebSearch *WebSearch `json:"web_search,omitempty"`
}

// Function describes a function for tool calling.
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Arguments   string                 `json:"arguments,omitempty"`
}

// WebSearch configures web search settings.
type WebSearch struct {
	SearchEngine        string `json:"search_engine"`
	SearchRecencyFilter string `json:"search_recency_filter,omitempty"`
	Count               int    `json:"count"`
	SearchIntent        bool   `json:"search_intent"`
	SearchDomainFilter  string `json:"search_domain_filter,omitempty"`
	ContentSize         string `json:"content_size,omitempty"`
}

// ToolCall represents a function call request from the assistant.
type ToolCall struct {
	ID       string    `json:"id,omitempty"`
	Type     string    `json:"type"`
	Index    int       `json:"index"`
	Function *Function `json:"function"`
}

// StreamEvent represents an event from the SSE stream.
type StreamEvent struct {
	Event    string
	Think    string
	Text     string
	ToolCall *ToolCall
	Raw      map[string]interface{}
	Error    error
}

// Option is a function that configures a Request.
type Option func(*Request)

// WithSystemPrompt sets the system prompt for the conversation.
// The system prompt defines the AI's behavior and personality.
//
// Example:
//
//	client.Chat(ctx, "Hello",
//	    reZ.WithSystemPrompt("You are a helpful coding assistant"))
func WithSystemPrompt(prompt string) Option {
	return func(r *Request) {
		r.SystemPrompt = prompt
	}
}

// WithTemperature sets the sampling temperature (0.0 to 2.0).
// Higher values make output more random, lower values more deterministic.
// Default is 1.0.
// Returns an error via panic if temperature is out of range.
//
// Example:
//
//	client.Chat(ctx, "Write a story",
//	    reZ.WithTemperature(0.8)) // More creative
func WithTemperature(temp float64) Option {
	if temp < 0.0 || temp > 2.0 {
		panic(fmt.Sprintf("temperature must be between 0.0 and 2.0, got: %.2f", temp))
	}
	return func(r *Request) {
		r.Temperature = temp
	}
}

// WithMaxTokens sets the maximum number of tokens in the response.
// Default is 65536. Must be greater than 0.
//
// Example:
//
//	client.Chat(ctx, "Summarize this",
//	    reZ.WithMaxTokens(500)) // Short response
func WithMaxTokens(tokens int) Option {
	if tokens <= 0 {
		panic(fmt.Sprintf("max_tokens must be greater than 0, got: %d", tokens))
	}
	return func(r *Request) {
		r.MaxTokens = tokens
	}
}

// WithThinking enables or disables the thinking mode.
// When enabled, the AI shows its reasoning process via StreamEvent.Think.
// Default is enabled.
//
// Example:
//
//	client.Chat(ctx, "Solve this puzzle",
//	    reZ.WithThinking(true))
func WithThinking(enabled bool) Option {
	return func(r *Request) {
		if enabled {
			r.Thinking = &Thinking{Type: "enabled"}
		} else {
			r.Thinking = nil
		}
	}
}

// WithWebSearch enables web search functionality.
// The AI can search the internet for real-time information.
//
// Example:
//
//	client.Chat(ctx, "Latest tech news",
//	    reZ.WithWebSearch(
//	        reZ.WithSearchRecency("oneDay"),
//	        reZ.WithSearchIntent(true),
//	    ))
func WithWebSearch(opts ...WebSearchOption) Option {
	return func(r *Request) {
		ws := &WebSearch{
			SearchEngine:        "search_std",
			SearchRecencyFilter: "noLimit",
			Count:               10,
			SearchIntent:        false,
			ContentSize:         "medium",
		}
		
		for _, opt := range opts {
			opt(ws)
		}
		
		r.Tools = append(r.Tools, Tool{
			Type:      "web_search",
			WebSearch: ws,
		})
	}
}

// WithFunction adds a function that the AI can call.
// The AI will request to call this function when appropriate.
//
// Parameters:
//   - name: Function name
//   - description: What the function does
//   - parameters: JSON schema for function parameters
//
// Example:
//
//	params := map[string]interface{}{
//	    "type": "object",
//	    "properties": map[string]interface{}{
//	        "location": map[string]interface{}{
//	            "type": "string",
//	            "description": "City name",
//	        },
//	    },
//	    "required": []string{"location"},
//	}
//	client.ChatWithHistory(ctx, "What's the weather?",
//	    reZ.WithFunction("get_weather", "Get weather", params))
func WithFunction(name, description string, parameters map[string]interface{}) Option {
	return func(r *Request) {
		r.Tools = append(r.Tools, Tool{
			Type: "function",
			Function: &Function{
				Name:        name,
				Description: description,
				Parameters:  parameters,
			},
		})
	}
}

// WebSearchOption is a function that configures WebSearch settings.
type WebSearchOption func(*WebSearch)

// WithSearchRecency sets the time filter for search results.
// Valid values: "noLimit", "oneDay", "oneWeek"
// Default is "noLimit".
func WithSearchRecency(filter string) WebSearchOption {
	validFilters := map[string]bool{
		"noLimit": true,
		"oneDay":  true,
		"oneWeek": true,
	}
	if !validFilters[filter] {
		panic(fmt.Sprintf("invalid search recency filter: %s (valid: noLimit, oneDay, oneWeek)", filter))
	}
	return func(ws *WebSearch) {
		ws.SearchRecencyFilter = filter
	}
}

// WithSearchDomain limits search to a specific domain.
// Example: "github.com", "stackoverflow.com"
func WithSearchDomain(domain string) WebSearchOption {
	return func(ws *WebSearch) {
		ws.SearchDomainFilter = domain
	}
}

// WithSearchIntent enables or disables search intent detection.
// When enabled, the AI better understands the search query.
// Default is false.
func WithSearchIntent(enabled bool) WebSearchOption {
	return func(ws *WebSearch) {
		ws.SearchIntent = enabled
	}
}

// WithSearchCount sets the number of search results to retrieve.
// Default is 10. Must be between 1 and 100.
func WithSearchCount(count int) WebSearchOption {
	if count < 1 || count > 100 {
		panic(fmt.Sprintf("search count must be between 1 and 100, got: %d", count))
	}
	return func(ws *WebSearch) {
		ws.Count = count
	}
}

// WithContentSize sets the amount of content to retrieve from each result.
// Valid values: "low", "medium", "high"
// Default is "medium".
func WithContentSize(size string) WebSearchOption {
	validSizes := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	if !validSizes[size] {
		panic(fmt.Sprintf("invalid content size: %s (valid: low, medium, high)", size))
	}
	return func(ws *WebSearch) {
		ws.ContentSize = size
	}
}
