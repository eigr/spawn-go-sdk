package spawn

import (
	"google.golang.org/protobuf/proto"
)

// Kind defines the type of the actor (e.g., UNNAMED, NAMED).
type Kind string

const (
	Unnamed    Kind = "UNNAMED"
	Named      Kind = "NAMED"
	Pooled     Kind = "Pooled"
	Task       Kind = "Task"
	Projection Kind = "Projection"
)

type Workflow struct{}

// Value represents the return on a stock.
type Value struct {
	State    proto.Message
	Response proto.Message
	Workflow interface{}
}

// ValueBuilder is the builder to create an instance of Value.
type ValueBuilder struct {
	value Value
}

// Of creates a new ValueBuilder with the initial response.
func Of(response proto.Message) *ValueBuilder {
	return &ValueBuilder{
		value: Value{Response: response},
	}
}

// State defines the state.
func (b *ValueBuilder) State(state proto.Message) *ValueBuilder {
	b.value.State = state
	return b
}

// Workflow creates a new flow.
func (b *ValueBuilder) Workflow(workflow interface{}) *ValueBuilder {
	b.value.Workflow = workflow
	return b
}

// Materialize finalizes the builder and returns the constructed Value.
func (b *ValueBuilder) Materialize() Value {
	return b.value
}

// ActorContext provides context for an actor's handler.
type ActorContext struct {
	CurrentState proto.Message
}
