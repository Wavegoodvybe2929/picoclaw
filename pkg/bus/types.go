package bus

type InboundMessage struct {
	Channel    string            `json:"channel"`
	SenderID   string            `json:"sender_id"`
	ChatID     string            `json:"chat_id"`
	Content    string            `json:"content"`
	Media      []string          `json:"media,omitempty"`
	SessionKey string            `json:"session_key"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type OutboundMessage struct {
	Channel string `json:"channel"`
	ChatID  string `json:"chat_id"`
	Content string `json:"content"`
}

// InputRequest represents a request for user input sent by an agent hook.
// The agent will block waiting for the user's response or timeout.
type InputRequest struct {
	RequestID string // UUID for tracking this specific request
	Channel   string // Channel where the request originated
	ChatID    string // Chat ID where the request originated
	Prompt    string // Question/prompt to show the user
	Timeout   int    // Seconds to wait for response
}

// InputResponse represents the user's response to an InputRequest.
// If TimedOut is true, the Input field may be empty and default value should be used.
type InputResponse struct {
	RequestID string // UUID matching the InputRequest
	Input     string // User's response text
	TimedOut  bool   // True if the request expired without user response
}

type MessageHandler func(InboundMessage) error
