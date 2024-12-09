package spawn

// Kind defines the type of the actor (e.g., UNNAMED, NAMED).
type Kind string

const (
	Unnamed Kind = "UNNAMED"
	Named   Kind = "NAMED"
)

// Value represents the return on a stock.
type Value struct {
	State    interface{}
	Response interface{}
}

// Of creates a new Value.
func Of(state interface{}, response interface{}) Value {
	return Value{State: state, Response: response}
}

// ActorContext provides context for an actor's handler.
type ActorContext struct {
	CurrentState interface{}
}
