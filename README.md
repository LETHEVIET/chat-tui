# Chat TUI

A terminal-based chat interface for interacting with OpenAI-compatible LLM APIs with streaming support and detailed statistics.

## Features

- **OpenAI-Compatible API Support** - Works with OpenAI, Ollama, and any OpenAI-compatible endpoint
- **Real-time Streaming** - Character-by-character streaming responses
- **Detailed Performance Stats** -
  - Time to First Token (TTFT)
  - Generation speed (tokens/sec)
  - Post-first-token generation speed
  - Total latency and token counts
- **Markdown Rendering** - Beautifully rendered markdown with syntax highlighting
- **Slash Commands** - Quick actions via `/` commands
- **Configuration Management** - Easy config reload without restarting
- **Conversation Management** - Save, load, and manage chat sessions

## Installation

```bash
go build -o chat-tui
```

## Configuration

Create a `.chat-tui.yaml` file in your current directory or home directory:

```yaml
# Chat TUI Configuration
# OpenAI-compatible API settings
api_key: "your-api-key-here"  # Or use OPENAI_API_KEY environment variable
base_url: "https://api.openai.com/v1"  # Change for Ollama: http://localhost:11434/v1
model: "gpt-4"  # Or for Ollama: llama2, codellama, etc.
temperature: 0.7
max_tokens: 4096
system_prompt: "You are a helpful assistant"

ui:
  theme: dark
  show_stats: true
  syntax_highlight: true

debug:
  verbose: false
  log_file: .chat-tui.log
```

### Using with Ollama

To use with Ollama, update your config:

```yaml
api_key: "ollama"  # Ollama doesn't require a real API key
base_url: "http://localhost:11434/v1"
model: "llama2"  # or mistral, codellama, etc.
```

## Usage

### Basic Usage

```bash
# Use config from .chat-tui.yaml
./chat-tui

# Override model
./chat-tui --model gpt-4-turbo

# Override base URL
./chat-tui --base-url http://localhost:11434/v1 --model llama2

# Disable stats panel
./chat-tui --no-stats
```

### Keyboard Shortcuts

- `Enter` - Send message (or newline in multiline mode)
- `Ctrl+D` - Toggle multiline mode
- `Ctrl+C` - Cancel streaming / Exit
- `Ctrl+S` - Toggle stats panel
- `Ctrl+K` - Scroll up
- `Ctrl+J` - Scroll down

### Slash Commands

All available commands:

```
/help           - Show all commands
/new            - Start a new conversation
/clear          - Clear chat history (alias for /new)
/reload         - Reload configuration from .chat-tui.yaml
/temp <0-1>     - Set temperature (e.g., /temp 0.7)
/system <text>  - Set system prompt
/delete         - Delete last turn (user message + assistant response)
/save <file>    - Save conversation to file
/load <file>    - Load conversation from file
/tokens         - Show token usage
/cost           - Show estimated cost
/export         - Export conversation as markdown
/stats          - Toggle stats panel
/debug          - Toggle debug mode
/retry          - Retry last message
/copy           - Copy last response to clipboard
/edit           - Edit last message
/multiline      - Toggle multiline input mode
/exit           - Exit the application
```

## Performance Metrics

The stats panel shows detailed performance metrics:

- **Total Latency** - Full request-to-completion time
- **Time to 1st Token (TTFT)** - How long until the first token arrives (important for perceived responsiveness)
- **Generation Time** - Time from first token to last token
- **Avg Speed** - Overall tokens/second (including TTFT overhead)
- **Gen Speed** - Pure generation speed after first token (actual model throughput)

## Examples

### Quick Chat with OpenAI

```bash
export OPENAI_API_KEY="sk-..."
./chat-tui --model gpt-4
```

### Chat with Local Ollama

```bash
./chat-tui --base-url http://localhost:11434/v1 --model llama2
```

### Start with Custom System Prompt

Edit `.chat-tui.yaml`:
```yaml
system_prompt: "You are an expert Go programmer. Provide concise, idiomatic Go code examples."
```

Then run:
```bash
./chat-tui
```

### Switch Between Providers

You can reload configuration on-the-fly:

1. Edit `.chat-tui.yaml` to switch from OpenAI to Ollama
2. In the chat, type `/reload`
3. Continue chatting with the new provider

## Project Structure

```
chat-tui/
├── cmd/
│   └── root.go          # Cobra CLI commands
├── internal/
│   ├── ui/
│   │   ├── chat.go      # Main Bubble Tea model
│   │   ├── components/  # Reusable UI components
│   │   │   ├── message.go
│   │   │   ├── input.go
│   │   │   └── stats.go
│   │   └── styles.go    # Lipgloss styles
│   ├── llm/
│   │   ├── client.go    # LLM client interface
│   │   └── openai.go    # OpenAI-compatible implementation
│   ├── config/
│   │   └── config.go    # Configuration management
│   └── commands/
│       └── slash.go     # Slash command parser
├── go.mod
├── main.go
└── README.md
```

## License

MIT
