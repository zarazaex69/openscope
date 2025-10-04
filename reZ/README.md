# reZ - Reverse Zhipu AI Client

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/derx/openscope/reZ)](https://goreportcard.com/report/github.com/derx/openscope/reZ)
[![Documentation](https://img.shields.io/badge/docs-godoc-blue.svg)](https://pkg.go.dev/github.com/derx/openscope/reZ)
[![GitHub stars](https://img.shields.io/github/stars/derx/openscope?style=social)](https://github.com/derx/openscope)

🔄 Go library for working with Zhipu AI API (GLM-4.6) through reverse engineering.

---

## 📑 Table of Contents

- [⚡ Features](#-features)
- [📦 Installation](#-installation)
- [🚀 Quick Start](#-quick-start)
  - [Simple Request](#simple-request)
  - [Conversation with History](#conversation-with-history)
  - [Web Search](#web-search)
  - [Function Calling](#function-calling)
- [📖 API Reference](#-api-reference)
  - [Client Methods](#client-methods)
  - [Configuration Options](#configuration-options)
  - [Types](#types)
- [💡 Advanced Usage](#-advanced-usage)
  - [Context Management](#context-management)
  - [Error Handling](#error-handling)
  - [Thinking Mode](#thinking-mode)
- [🎯 Best Practices](#-best-practices)
- [⚠️ Limitations](#️-limitations)
- [📄 License](#-license)

---

## ⚡ Features

[↑ Back to top](#-table-of-contents)

- ✅ **Streaming output** (SSE streaming)
- ✅ **Conversation history** (automatic context management)
- ✅ **Thinking mode** (shows AI reasoning)
- ✅ **Web search** (real-time information)
- ✅ **Tool Calling** (function calling)
- ✅ **System prompts**
- ✅ **Flexible configuration** (temperature, max_tokens, top_p)

## 📦 Installation

[↑ Back to top](#-table-of-contents)

```bash
go get github.com/zarazaex69/openscope/reZ
```

Or install locally:

```bash
cd reZ
go install
```

## 🚀 Quick Start

[↑ Back to top](#-table-of-contents)

### Simple Request

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/zarazaex69/openscope/reZ"
)

func main() {
    client := reZ.NewClient()
    ctx := context.Background()
    
    events, err := client.Chat(ctx, "Hello!")
    if err != nil {
        log.Fatal(err)
    }
    
    for event := range events {
        if event.Error != nil {
            log.Fatal(event.Error)
        }
        
        if event.Text != "" {
            fmt.Print(event.Text)
        }
    }
}
```

### With Conversation History

```go
client := reZ.NewClient()

// First message
events, _ := client.ChatWithHistory(ctx, "My name is Alex")
for event := range events {
    fmt.Print(event.Text)
}

// Second message (with context)
events, _ = client.ChatWithHistory(ctx, "What's my name?")
for event := range events {
    fmt.Print(event.Text) // "Your name is Alex"
}

// Get history
history := client.GetHistory()
```

### With Web Search

```go
events, err := client.Chat(
    ctx,
    "What's the latest news?",
    reZ.WithWebSearch(
        reZ.WithSearchRecency("oneDay"),    // Last 24 hours
        reZ.WithSearchIntent(true),         // Intent detection
        reZ.WithContentSize("high"),        // More content
    ),
)
```

### Tool Calling (Function Calling)

```go
// Define function
weatherParams := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "location": map[string]interface{}{
            "type": "string",
            "description": "City name",
        },
    },
    "required": []string{"location"},
}

// Request with function
events, _ := client.ChatWithHistory(
    ctx,
    "What's the weather in Moscow?",
    reZ.WithFunction("get_weather", "Get weather", weatherParams),
)

for event := range events {
    if event.ToolCall != nil {
        // AI requested function call
        fmt.Printf("Calling: %s(%s)\n", 
            event.ToolCall.Function.Name,
            event.ToolCall.Function.Arguments)
        
        // Execute function
        result := getWeather("Moscow")
        
        // Return result
        client.AddToolResponse(result)
        
        // Continue conversation
        events2, _ := client.ChatWithHistory(ctx, "")
        // ...
    }
}
```

## 🎛 Options

[↑ Back to top](#-table-of-contents)

### Basic Settings

```go
reZ.WithSystemPrompt("You are a helpful assistant")
reZ.WithTemperature(0.7)
reZ.WithMaxTokens(4096)
reZ.WithThinking(false)  // Disable thinking mode
```

### Web Search

```go
reZ.WithWebSearch(
    reZ.WithSearchRecency("oneDay"),      // "noLimit", "oneDay", "oneWeek"
    reZ.WithSearchDomain("github.com"),   // Search specific domain
    reZ.WithSearchIntent(true),           // Intent detection
    reZ.WithSearchCount(20),              // Number of results
    reZ.WithContentSize("high"),          // "low", "medium", "high"
)
```

### Tool Calling

```go
reZ.WithFunction(
    "function_name",
    "Function description",
    map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param1"},
    },
)
```

## 📊 StreamEvent

[↑ Back to top](#-table-of-contents)

Each event from the stream contains:

```go
type StreamEvent struct {
    Event    string                 // Event type
    Think    string                 // Thinking process text
    Text     string                 // Response text
    ToolCall *ToolCall              // Function call (if any)
    Raw      map[string]interface{} // Raw data
    Error    error                  // Error (if any)
}
```

## 🔧 Quick API Reference

[↑ Back to top](#-table-of-contents)

### Client

```go
// Create client
client := reZ.NewClient()

// Simple request
Chat(ctx, content, ...opts) (<-chan StreamEvent, error)

// Request with history
ChatWithHistory(ctx, content, ...opts) (<-chan StreamEvent, error)

// Add tool response
AddToolResponse(content)

// Clear history
ClearHistory()

// Get history
GetHistory() []Message
```

## 📝 More Examples

[↑ Back to top](#-table-of-contents)

See `examples/` folder:
- `basic/` - simple request
- `history/` - working with history
- `websearch/` - web search
- `toolcalling/` - function calling

## 📖 Full Documentation

[↑ Back to top](#-table-of-contents)

View documentation locally:

```bash
go doc -all github.com/zarazaex69/openscope/reZ
```

## 💡 Advanced Usage

[↑ Back to top](#-table-of-contents)

### Context Management

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

events, err := client.Chat(ctx, "Long task")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Request timed out")
    }
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel after 5 seconds
}()

events, _ := client.Chat(ctx, "Task")
```

### Error Handling

Always check for errors in stream events:

```go
events, err := client.Chat(ctx, "Hello")
if err != nil {
    log.Fatalf("Failed to start chat: %v", err)
}

for event := range events {
    if event.Error != nil {
        log.Printf("Stream error: %v", event.Error)
        continue
    }
    
    fmt.Print(event.Text)
}
```

### Thinking Mode

View the AI's reasoning process:

```go
events, _ := client.Chat(ctx, "Solve this puzzle: ...",
    reZ.WithThinking(true))

for event := range events {
    if event.Think != "" {
        fmt.Printf("[THINKING]: %s", event.Think)
    }
    if event.Text != "" {
        fmt.Printf("[RESPONSE]: %s", event.Text)
    }
}
```

## 📖 API Reference

[↑ Back to top](#-table-of-contents)


### Client Methods

- `NewClient() *Client` - Create new client
- `Chat(ctx, content, ...opts) (<-chan StreamEvent, error)` - Simple request
- `ChatWithHistory(ctx, content, ...opts) (<-chan StreamEvent, error)` - Request with history
- `AddToolResponse(content)` - Add tool response
- `ClearHistory()` - Clear history
- `GetHistory() []Message` - Get history

### Configuration Options

- `WithSystemPrompt(prompt)` - Set system prompt
- `WithTemperature(temp)` - Set temperature (0.0-2.0)
- `WithMaxTokens(tokens)` - Set max tokens
- `WithThinking(enabled)` - Enable/disable thinking mode
- `WithWebSearch(...opts)` - Enable web search
- `WithFunction(name, desc, params)` - Add function

### Types

- `StreamEvent` - Event from stream
- `Message` - Conversation message
- `ToolCall` - Function call request
- `Function` - Function details

---

## 📚 Examples

[↑ Back to top](#-table-of-contents)

### Example 1: Simple Chat

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/zarazaex69/openscope/reZ"
)

func main() {
    client := reZ.NewClient()
    ctx := context.Background()
    
    events, err := client.Chat(ctx, "Tell me a joke")
    if err != nil {
        log.Fatal(err)
    }
    
    for event := range events {
        if event.Error != nil {
            log.Fatal(event.Error)
        }
        fmt.Print(event.Text)
    }
}
```

### Example 2: Conversation with History

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/zarazaex69/openscope/reZ"
)

func main() {
    client := reZ.NewClient()
    ctx := context.Background()
    
    messages := []string{
        "My favorite color is blue",
        "What's my favorite color?",
    }
    
    for _, msg := range messages {
        fmt.Printf("User: %s\n", msg)
        
        events, err := client.ChatWithHistory(ctx, msg)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Print("AI: ")
        for event := range events {
            if event.Error != nil {
                log.Fatal(event.Error)
            }
            fmt.Print(event.Text)
        }
        fmt.Println()
    }
}
```

### Example 3: Web Search

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/zarazaex69/openscope/reZ"
)

func main() {
    client := reZ.NewClient()
    ctx := context.Background()
    
    events, err := client.Chat(ctx, 
        "What are the latest developments in AI?",
        reZ.WithWebSearch(
            reZ.WithSearchRecency("oneDay"),
            reZ.WithSearchIntent(true),
        ))
    
    if err != nil {
        log.Fatal(err)
    }
    
    for event := range events {
        if event.Error != nil {
            log.Fatal(event.Error)
        }
        fmt.Print(event.Text)
    }
}
```

### Example 4: Function Calling

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "github.com/zarazaex69/openscope/reZ"
)

func getWeather(location string) string {
    // Simulate API call
    return fmt.Sprintf(`{"location": "%s", "temp": 15, "condition": "cloudy"}`, location)
}

func main() {
    client := reZ.NewClient()
    ctx := context.Background()
    
    params := map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type": "string",
                "description": "City name",
            },
        },
        "required": []string{"location"},
    }
    
    events, _ := client.ChatWithHistory(ctx, 
        "What's the weather in London?",
        reZ.WithFunction("get_weather", "Get weather", params))
    
    for event := range events {
        if event.ToolCall != nil {
            var args map[string]interface{}
            json.Unmarshal([]byte(event.ToolCall.Function.Arguments), &args)
            
            result := getWeather(args["location"].(string))
            client.AddToolResponse(result)
            
            events2, _ := client.ChatWithHistory(ctx, "",
                reZ.WithFunction("get_weather", "Get weather", params))
            
            for e := range events2 {
                fmt.Print(e.Text)
            }
        }
    }
}
```

---

## 🎯 Best Practices

[↑ Back to top](#-table-of-contents)

### 1. Always Handle Errors

```go
events, err := client.Chat(ctx, "Hello")
if err != nil {
    log.Fatal(err)
}

for event := range events {
    if event.Error != nil {
        log.Printf("Error: %v", event.Error)
        continue
    }
    // Process event
}
```

### 2. Use Context for Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

events, err := client.Chat(ctx, "Long task")
```

### 3. Clear History When Needed

```go
// After completing a conversation
client.ClearHistory()

// Start new conversation
client.ChatWithHistory(ctx, "New topic")
```

### 4. Buffer Responses

```go
var response strings.Builder

for event := range events {
    if event.Text != "" {
        response.WriteString(event.Text)
    }
}

fullResponse := response.String()
```

### 5. Separate Thinking from Response

```go
var thinking, response strings.Builder

for event := range events {
    if event.Think != "" {
        thinking.WriteString(event.Think)
    }
    if event.Text != "" {
        response.WriteString(event.Text)
    }
}
```

---

## ⚠️ Limitations

[↑ Back to top](#-table-of-contents)

- **Hardcoded credentials** - For educational purposes only
- **No file upload** - File attachments not yet supported
- **Single model** - Only GLM-4.6 supported
- **Rate limits** - Depends on API endpoint
- **Thread-safe** - Client can be used concurrently

---

## 📄 License

[↑ Back to top](#-table-of-contents)

MIT License - see [LICENSE](LICENSE) for details.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## ⭐ Show your support

Give a ⭐️ if this project helped you!

---

**Made with ❤️ for the Go community**
