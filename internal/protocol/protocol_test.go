package protocol

import (
	"encoding/json"
	"testing"
)

// ---------------------------------------------------------------------------
// State Machine Tests
// ---------------------------------------------------------------------------

func TestCanTransition_ValidTransitions(t *testing.T) {
	valid := []struct {
		from, to AgentState
	}{
		{StateSpawning, StateReady},
		{StateSpawning, StateError},
		{StateReady, StateIdle},
		{StateIdle, StateBusy},
		{StateIdle, StateError},
		{StateIdle, StateShutdown},
		{StateBusy, StateCompleted},
		{StateBusy, StateError},
		{StateCompleted, StateIdle},
		{StateError, StateIdle},
		{StateError, StateShutdown},
	}

	for _, tc := range valid {
		if !CanTransition(tc.from, tc.to) {
			t.Errorf("expected %s → %s to be valid", tc.from, tc.to)
		}
	}
}

func TestCanTransition_InvalidTransitions(t *testing.T) {
	invalid := []struct {
		from, to AgentState
	}{
		{StateShutdown, StateIdle},     // terminal, no way out
		{StateShutdown, StateSpawning}, // terminal
		{StateIdle, StateReady},        // backwards
		{StateBusy, StateIdle},         // must go through completed
		{StateSpawning, StateBusy},     // must go through ready→idle first
		{StateReady, StateBusy},        // must go through idle
		{StateCompleted, StateShutdown},// must go through idle
	}

	for _, tc := range invalid {
		if CanTransition(tc.from, tc.to) {
			t.Errorf("expected %s → %s to be invalid", tc.from, tc.to)
		}
	}
}

func TestAgentState_IsTerminal(t *testing.T) {
	if !StateShutdown.IsTerminal() {
		t.Error("shutdown should be terminal")
	}
	if StateIdle.IsTerminal() {
		t.Error("idle should not be terminal")
	}
}

func TestAgentState_IsTransient(t *testing.T) {
	if !StateReady.IsTransient() {
		t.Error("ready should be transient")
	}
	if !StateCompleted.IsTransient() {
		t.Error("completed should be transient")
	}
	if StateIdle.IsTransient() {
		t.Error("idle should not be transient")
	}
}

// ---------------------------------------------------------------------------
// Message Tests
// ---------------------------------------------------------------------------

func TestNewMessage(t *testing.T) {
	msg := NewMessage(TypeMessage, "agent-a", "agent-b", TextPayload("hello"))

	if msg.ID == "" {
		t.Error("message ID should not be empty")
	}
	if msg.Type != TypeMessage {
		t.Errorf("expected type %s, got %s", TypeMessage, msg.Type)
	}
	if msg.From != "agent-a" {
		t.Errorf("expected from agent-a, got %s", msg.From)
	}
	if msg.To != "agent-b" {
		t.Errorf("expected to agent-b, got %s", msg.To)
	}
	if msg.Payload.Content != "hello" {
		t.Errorf("expected content hello, got %s", msg.Payload.Content)
	}
	if msg.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestNewRequest(t *testing.T) {
	msg := NewRequest("a", "b", TextPayload("need info"))
	if msg.Type != TypeRequest {
		t.Errorf("expected TypeRequest, got %s", msg.Type)
	}
}

func TestNewResponse(t *testing.T) {
	msg := NewResponse("b", "a", "corr-123", TextPayload("here you go"))
	if msg.Type != TypeResponse {
		t.Errorf("expected TypeResponse, got %s", msg.Type)
	}
	if msg.CorrelationID != "corr-123" {
		t.Errorf("expected correlation_id corr-123, got %s", msg.CorrelationID)
	}
}

func TestNewBroadcast(t *testing.T) {
	msg := NewBroadcast("a", TextPayload("attention everyone"))
	if msg.Type != TypeBroadcast {
		t.Errorf("expected TypeBroadcast, got %s", msg.Type)
	}
	if msg.To != BroadcastTarget {
		t.Errorf("expected to=%s, got %s", BroadcastTarget, msg.To)
	}
}

func TestJSONPayload(t *testing.T) {
	type data struct {
		Key string `json:"key"`
	}
	p, err := JSONPayload(data{Key: "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ContentType != ContentJSON {
		t.Errorf("expected ContentJSON, got %s", p.ContentType)
	}
	if p.Content != `{"key":"value"}` {
		t.Errorf("unexpected content: %s", p.Content)
	}
}

// ---------------------------------------------------------------------------
// Message Validation Tests
// ---------------------------------------------------------------------------

func TestMessage_Validate_Valid(t *testing.T) {
	msg := NewMessage(TypeMessage, "a", "b", TextPayload("hi"))
	if err := msg.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestMessage_Validate_BroadcastValid(t *testing.T) {
	msg := NewBroadcast("a", TextPayload("hi all"))
	if err := msg.Validate(); err != nil {
		t.Errorf("expected valid broadcast, got error: %v", err)
	}
}

func TestMessage_Validate_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		msg  Message
	}{
		{"missing id", Message{Type: TypeMessage, From: "a", To: "b", Payload: TextPayload("x")}},
		{"missing type", Message{ID: "1", From: "a", To: "b", Payload: TextPayload("x")}},
		{"missing from", Message{ID: "1", Type: TypeMessage, To: "b", Payload: TextPayload("x")}},
		{"missing to", Message{ID: "1", Type: TypeMessage, From: "a", Payload: TextPayload("x")}},
	}

	for _, tc := range cases {
		if err := tc.msg.Validate(); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

func TestMessage_Validate_BroadcastWrongTarget(t *testing.T) {
	msg := NewMessage(TypeBroadcast, "a", "b", TextPayload("x"))
	if err := msg.Validate(); err == nil {
		t.Error("broadcast with non-* target should fail validation")
	}
}

func TestMessage_Validate_NonBroadcastWithStar(t *testing.T) {
	msg := NewMessage(TypeMessage, "a", BroadcastTarget, TextPayload("x"))
	if err := msg.Validate(); err == nil {
		t.Error("non-broadcast with * target should fail validation")
	}
}

func TestMessage_Validate_ResponseWithoutCorrelation(t *testing.T) {
	msg := NewMessage(TypeResponse, "a", "b", TextPayload("x"))
	if err := msg.Validate(); err == nil {
		t.Error("response without correlation_id should fail validation")
	}
}

func TestMessage_Validate_PayloadTooLarge(t *testing.T) {
	bigContent := make([]byte, MaxPayloadSize+1)
	for i := range bigContent {
		bigContent[i] = 'x'
	}
	msg := NewMessage(TypeMessage, "a", "b", Payload{ContentType: ContentText, Content: string(bigContent)})
	if err := msg.Validate(); err == nil {
		t.Error("oversized payload should fail validation")
	}
}

// ---------------------------------------------------------------------------
// JSON Serialization Round-trip
// ---------------------------------------------------------------------------

func TestMessage_JSONRoundTrip(t *testing.T) {
	original := NewMessage(TypeMessage, "agent-a", "agent-b", TextPayload("hello world"))
	original.Metadata = map[string]any{"priority": "high"}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", decoded.ID, original.ID)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type mismatch: %s vs %s", decoded.Type, original.Type)
	}
	if decoded.From != original.From {
		t.Errorf("From mismatch")
	}
	if decoded.To != original.To {
		t.Errorf("To mismatch")
	}
	if decoded.Payload.Content != original.Payload.Content {
		t.Errorf("Content mismatch")
	}
	if decoded.Payload.ContentType != original.Payload.ContentType {
		t.Errorf("ContentType mismatch")
	}
}

// ---------------------------------------------------------------------------
// Event Tests
// ---------------------------------------------------------------------------

func TestNewAgentSpawnedEvent(t *testing.T) {
	evt := NewAgentSpawnedEvent("agent-1", "researcher", "Research Agent", []string{"web_search", "summarize"})

	if evt.ID == "" {
		t.Error("event ID should not be empty")
	}
	if evt.Type != EventAgentSpawned {
		t.Errorf("expected type %s, got %s", EventAgentSpawned, evt.Type)
	}
	if evt.AgentID != "agent-1" {
		t.Errorf("expected agent_id agent-1, got %s", evt.AgentID)
	}
	if evt.Sequence != 0 {
		t.Errorf("sequence should be 0 before assignment, got %d", evt.Sequence)
	}

	data, ok := evt.Data.(AgentSpawnedData)
	if !ok {
		t.Fatalf("data should be AgentSpawnedData, got %T", evt.Data)
	}
	if data.Name != "researcher" {
		t.Errorf("expected name researcher, got %s", data.Name)
	}
	if len(data.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(data.Capabilities))
	}
}

func TestNewAgentStateChangedEvent(t *testing.T) {
	evt := NewAgentStateChangedEvent("agent-1", StateIdle, StateBusy, "received task")
	data, ok := evt.Data.(AgentStateChangedData)
	if !ok {
		t.Fatal("data should be AgentStateChangedData")
	}
	if data.PrevState != StateIdle || data.NewState != StateBusy {
		t.Error("state mismatch")
	}
	if data.Reason != "received task" {
		t.Errorf("expected reason 'received task', got %s", data.Reason)
	}
}

func TestNewAgentMessageSentEvent(t *testing.T) {
	msg := NewMessage(TypeMessage, "a", "b", TextPayload("hi"))
	evt := NewAgentMessageSentEvent("a", msg)

	data, ok := evt.Data.(AgentMessageSentData)
	if !ok {
		t.Fatal("data should be AgentMessageSentData")
	}
	if data.Message.ID != msg.ID {
		t.Error("message ID mismatch")
	}
}

func TestNewAgentErrorEvent(t *testing.T) {
	evt := NewAgentErrorEvent("agent-1", "connection refused", ErrorUndeliverable, true)
	data, ok := evt.Data.(AgentErrorData)
	if !ok {
		t.Fatal("data should be AgentErrorData")
	}
	if data.Kind != ErrorUndeliverable {
		t.Errorf("expected kind %s, got %s", ErrorUndeliverable, data.Kind)
	}
	if !data.Recoverable {
		t.Error("expected recoverable=true")
	}
}

func TestNewTopologyChangedEvent(t *testing.T) {
	evt := NewTopologyChangedEvent(TopologyAgentJoined, map[string]any{"agent_id": "agent-1"})
	if evt.AgentID != "" {
		t.Error("topology events should have empty agent_id")
	}
	data, ok := evt.Data.(TopologyChangedData)
	if !ok {
		t.Fatal("data should be TopologyChangedData")
	}
	if data.ChangeType != TopologyAgentJoined {
		t.Errorf("expected change_type %s, got %s", TopologyAgentJoined, data.ChangeType)
	}
}

func TestEvent_JSONRoundTrip(t *testing.T) {
	evt := NewAgentSpawnedEvent("agent-1", "researcher", "Research Agent", []string{"search"})
	evt.Sequence = 42

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != evt.ID {
		t.Errorf("ID mismatch")
	}
	if decoded.Type != evt.Type {
		t.Errorf("Type mismatch")
	}
	if decoded.Sequence != 42 {
		t.Errorf("Sequence mismatch: expected 42, got %d", decoded.Sequence)
	}
	if decoded.AgentID != evt.AgentID {
		t.Errorf("AgentID mismatch")
	}
}
