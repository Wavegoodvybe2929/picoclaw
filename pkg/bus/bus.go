package bus

import (
	"context"
	"sync"
)

type MessageBus struct {
	inbound          chan InboundMessage
	outbound         chan OutboundMessage
	handlers         map[string]MessageHandler
	inputRequests    chan InputRequest
	inputSubscribers map[string]chan InputResponse
	closed           bool
	mu               sync.RWMutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inbound:          make(chan InboundMessage, 100),
		outbound:         make(chan OutboundMessage, 100),
		handlers:         make(map[string]MessageHandler),
		inputRequests:    make(chan InputRequest, 10),
		inputSubscribers: make(map[string]chan InputResponse),
	}
}

func (mb *MessageBus) PublishInbound(msg InboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	mb.inbound <- msg
}

func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case msg := <-mb.inbound:
		return msg, true
	case <-ctx.Done():
		return InboundMessage{}, false
	}
}

func (mb *MessageBus) PublishOutbound(msg OutboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	mb.outbound <- msg
}

func (mb *MessageBus) SubscribeOutbound(ctx context.Context) (OutboundMessage, bool) {
	select {
	case msg := <-mb.outbound:
		return msg, true
	case <-ctx.Done():
		return OutboundMessage{}, false
	}
}

func (mb *MessageBus) RegisterHandler(channel string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.handlers[channel] = handler
}

func (mb *MessageBus) GetHandler(channel string) (MessageHandler, bool) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	handler, ok := mb.handlers[channel]
	return handler, ok
}

func (mb *MessageBus) Close() {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if mb.closed {
		return
	}
	mb.closed = true
	close(mb.inbound)
	close(mb.outbound)
	close(mb.inputRequests)
	// Close all input subscriber channels
	for _, ch := range mb.inputSubscribers {
		close(ch)
	}
}

// PublishInputRequest sends an input request to be routed to the user.
// The request will be sent via the outbound channel with a formatted prompt.
func (mb *MessageBus) PublishInputRequest(req InputRequest) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	// Send the prompt to the user via outbound channel
	mb.outbound <- OutboundMessage{
		Channel: req.Channel,
		ChatID:  req.ChatID,
		Content: req.Prompt,
	}
	// Store the request for tracking
	mb.inputRequests <- req
}

// SubscribeInputResponse creates a channel to receive the response for a specific request ID.
// The caller should listen on this channel with a timeout to handle case where user doesn't respond.
func (mb *MessageBus) SubscribeInputResponse(requestID string) <-chan InputResponse {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan InputResponse, 1)
	mb.inputSubscribers[requestID] = ch
	return ch
}

// PublishInputResponse sends a user's input response for a specific request.
// This should be called when the user responds to an input request.
func (mb *MessageBus) PublishInputResponse(response InputResponse) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if ch, ok := mb.inputSubscribers[response.RequestID]; ok {
		select {
		case ch <- response:
			// Successfully sent response
		default:
			// Channel full or closed, ignore
		}
		// Clean up the subscriber
		close(ch)
		delete(mb.inputSubscribers, response.RequestID)
	}
}

// ConsumeInputRequest retrieves pending input requests.
// This is used by channel handlers to receive and display input prompts to users.
func (mb *MessageBus) ConsumeInputRequest(ctx context.Context) (InputRequest, bool) {
	select {
	case req := <-mb.inputRequests:
		return req, true
	case <-ctx.Done():
		return InputRequest{}, false
	}
}
