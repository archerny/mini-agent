package agent

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/archerny/mini-agent/internal/protocol"
)

// SendFunc is the callback the executor uses to send outbound messages.
// It is set by the Runtime/MessageBus when the agent is started.
type SendFunc func(msg *protocol.Message) error

// Start begins the agent's executor goroutine.
// The sendFn callback is used to route outbound messages through the MessageBus.
func (a *Agent) Start(ctx context.Context, sendFn SendFunc) error {
	// Transition: spawning → ready → idle
	if err := a.transition(protocol.StateReady, "initialized"); err != nil {
		return fmt.Errorf("start agent %s: %w", a.id, err)
	}
	// ready is transient → auto-transition to idle
	if err := a.transition(protocol.StateIdle, "ready"); err != nil {
		return fmt.Errorf("start agent %s: %w", a.id, err)
	}

	execCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()

	go a.run(execCtx, sendFn)
	return nil
}

// Shutdown requests the agent to stop processing.
func (a *Agent) Shutdown(reason string) error {
	state := a.State()
	// Only idle or error agents can be shut down.
	if state != protocol.StateIdle && state != protocol.StateError {
		return fmt.Errorf("agent %s: cannot shutdown from state %s", a.id, state)
	}
	if err := a.transition(protocol.StateShutdown, reason); err != nil {
		return err
	}
	a.mu.RLock()
	cancel := a.cancel
	a.mu.RUnlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

// run is the executor goroutine — the main message processing loop.
func (a *Agent) run(ctx context.Context, sendFn SendFunc) {
	defer close(a.done)
	defer func() {
		// If the goroutine exits unexpectedly (not via Shutdown), ensure we
		// transition to shutdown state.
		if a.State() != protocol.StateShutdown {
			_ = a.transition(protocol.StateShutdown, "executor exited")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-a.Inbox:
			if !ok {
				return // channel closed
			}
			a.processMessage(ctx, msg, sendFn)
		}
	}
}

// processMessage handles a single message with panic recovery.
func (a *Agent) processMessage(ctx context.Context, msg *protocol.Message, sendFn SendFunc) {
	// Transition: idle → busy
	if err := a.transition(protocol.StateBusy, fmt.Sprintf("processing message %s", msg.ID)); err != nil {
		log.Printf("[agent:%s] failed to transition to busy: %v", a.id, err)
		return
	}

	// Run handler with panic recovery
	var outMsgs []*protocol.Message
	var handlerErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				handlerErr = fmt.Errorf("panic in handler: %v\n%s", r, stack)
			}
		}()
		outMsgs, handlerErr = a.handler(ctx, a, msg)
	}()

	if handlerErr != nil {
		// Transition: busy → error
		_ = a.transition(protocol.StateError, handlerErr.Error())
		log.Printf("[agent:%s] handler error: %v", a.id, handlerErr)
		// Auto-recover: error → idle (MVP: always recoverable)
		_ = a.transition(protocol.StateIdle, "auto-recovered from error")
		return
	}

	// Transition: busy → completed → idle (completed is transient)
	if err := a.transition(protocol.StateCompleted, "message processed"); err != nil {
		log.Printf("[agent:%s] failed to transition to completed: %v", a.id, err)
		return
	}
	// completed is transient → auto-transition to idle
	if err := a.transition(protocol.StateIdle, "completed"); err != nil {
		log.Printf("[agent:%s] failed to transition to idle: %v", a.id, err)
		return
	}

	// Send outbound messages through the MessageBus
	for _, outMsg := range outMsgs {
		if err := sendFn(outMsg); err != nil {
			log.Printf("[agent:%s] failed to send message: %v", a.id, err)
		}
	}
}
