package bus

import (
	"context"
	"testing"
	"time"
)

func TestMessageBus_InputRequestPublishAndSubscribe(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	requestID := "test-request-123"
	testRequest := InputRequest{
		RequestID: requestID,
		Channel:   "telegram",
		ChatID:    "user123",
		Prompt:    "What is your favorite color?",
		Timeout:   60,
	}

	// Subscribe before publishing
	responseChan := mb.SubscribeInputResponse(requestID)

	// Publish the request
	mb.PublishInputRequest(testRequest)

	// Verify the prompt was sent via outbound channel
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	outMsg, ok := mb.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("Expected outbound message but got none")
	}

	if outMsg.Channel != testRequest.Channel {
		t.Errorf("Expected channel %s, got %s", testRequest.Channel, outMsg.Channel)
	}
	if outMsg.ChatID != testRequest.ChatID {
		t.Errorf("Expected chatID %s, got %s", testRequest.ChatID, outMsg.ChatID)
	}
	if outMsg.Content != testRequest.Prompt {
		t.Errorf("Expected content %s, got %s", testRequest.Prompt, outMsg.Content)
	}

	// Verify the request was queued
	reqReceived, ok := mb.ConsumeInputRequest(ctx)
	if !ok {
		t.Fatal("Expected input request but got none")
	}

	if reqReceived.RequestID != testRequest.RequestID {
		t.Errorf("Expected request ID %s, got %s", testRequest.RequestID, reqReceived.RequestID)
	}

	// Simulate user response
	testResponse := InputResponse{
		RequestID: requestID,
		Input:     "Blue",
		TimedOut:  false,
	}

	mb.PublishInputResponse(testResponse)

	// Verify subscriber received the response
	select {
	case resp := <-responseChan:
		if resp.RequestID != testResponse.RequestID {
			t.Errorf("Expected request ID %s, got %s", testResponse.RequestID, resp.RequestID)
		}
		if resp.Input != testResponse.Input {
			t.Errorf("Expected input %s, got %s", testResponse.Input, resp.Input)
		}
		if resp.TimedOut != testResponse.TimedOut {
			t.Errorf("Expected TimedOut %v, got %v", testResponse.TimedOut, resp.TimedOut)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for response")
	}
}

func TestMessageBus_InputResponseTimeout(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	requestID := "test-request-timeout"

	// Subscribe
	responseChan := mb.SubscribeInputResponse(requestID)

	// Send timeout response
	timeoutResponse := InputResponse{
		RequestID: requestID,
		Input:     "",
		TimedOut:  true,
	}

	mb.PublishInputResponse(timeoutResponse)

	// Verify timeout was received
	select {
	case resp := <-responseChan:
		if !resp.TimedOut {
			t.Error("Expected TimedOut to be true")
		}
		if resp.RequestID != requestID {
			t.Errorf("Expected request ID %s, got %s", requestID, resp.RequestID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for timeout response")
	}
}

func TestMessageBus_MultipleInputRequests(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	// Create multiple requests
	requestIDs := []string{"req1", "req2", "req3"}
	responses := make([]<-chan InputResponse, len(requestIDs))

	// Subscribe to all requests
	for i, id := range requestIDs {
		responses[i] = mb.SubscribeInputResponse(id)
	}

	// Publish responses in reverse order to test routing
	for i := len(requestIDs) - 1; i >= 0; i-- {
		mb.PublishInputResponse(InputResponse{
			RequestID: requestIDs[i],
			Input:     "Response " + requestIDs[i],
			TimedOut:  false,
		})
	}

	// Verify each subscriber got the correct response
	for i, responseChan := range responses {
		select {
		case resp := <-responseChan:
			expectedID := requestIDs[i]
			if resp.RequestID != expectedID {
				t.Errorf("Expected request ID %s, got %s", expectedID, resp.RequestID)
			}
			expectedInput := "Response " + expectedID
			if resp.Input != expectedInput {
				t.Errorf("Expected input %s, got %s", expectedInput, resp.Input)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Timeout waiting for response %d", i)
		}
	}
}

func TestMessageBus_PublishResponseWithoutSubscriber(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	// Publish response without subscriber - should not panic
	mb.PublishInputResponse(InputResponse{
		RequestID: "no-subscriber",
		Input:     "test",
		TimedOut:  false,
	})

	// Test passes if no panic occurs
}

func TestMessageBus_CloseWithPendingInputRequests(t *testing.T) {
	mb := NewMessageBus()

	requestID := "test-close"
	responseChan := mb.SubscribeInputResponse(requestID)

	// Close the bus
	mb.Close()

	// Verify subscriber channel is closed
	select {
	case _, ok := <-responseChan:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for channel close")
	}

	// Verify PublishInputRequest doesn't panic when closed
	mb.PublishInputRequest(InputRequest{
		RequestID: "after-close",
		Channel:   "test",
		ChatID:    "test",
		Prompt:    "test",
		Timeout:   60,
	})
}

func TestMessageBus_ConsumeInputRequestWithContext(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	testRequest := InputRequest{
		RequestID: "ctx-test",
		Channel:   "slack",
		ChatID:    "channel123",
		Prompt:    "Test prompt",
		Timeout:   30,
	}

	// Start goroutine to publish after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		mb.PublishInputRequest(testRequest)
	}()

	// Consume with context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, ok := mb.ConsumeInputRequest(ctx)
	if !ok {
		t.Fatal("Expected to consume input request")
	}

	if req.RequestID != testRequest.RequestID {
		t.Errorf("Expected request ID %s, got %s", testRequest.RequestID, req.RequestID)
	}
}

func TestMessageBus_ConsumeInputRequestContextCancellation(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to consume with cancelled context
	_, ok := mb.ConsumeInputRequest(ctx)
	if ok {
		t.Error("Expected ConsumeInputRequest to return false with cancelled context")
	}
}
