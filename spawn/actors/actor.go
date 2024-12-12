package actors

import (
	"sync"

	"google.golang.org/protobuf/proto"
)

// ActionHandler defines the function of a Protobuf-supported action.
type ActionHandler func(ctx *ActorContext, payload proto.Message) (Value, error)

// Actor represents an actor in Spawn.
type Actor struct {
	Name               string
	StateType          proto.Message
	Kind               Kind
	Stateful           bool
	SnapshotTimeout    int64
	DeactivatedTimeout int64
	MinPoolSize        int32
	MaxPoolSize        int32
	Actions            map[string]ActionHandler
	mu                 sync.Mutex
}

// ActorConfig configures an actor.
type ActorConfig struct {
	Name               string
	StateType          proto.Message
	Kind               Kind
	Stateful           bool
	SnapshotTimeout    int64
	DeactivatedTimeout int64
	MinPoolSize        int32
	MaxPoolSize        int32
}

// ActorOf creates a new actor instance (preferred method for API consistency).
func ActorOf(config ActorConfig) *Actor {
	return newActor(config) // Delegates to the existing constructor for simplicity.
}

// AddAction adds a new action to the actor.
func (a *Actor) AddAction(name string, handler ActionHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Actions[name] = handler
}

// NewActor creates a new actor instance (legacy method, can be deprecated if needed).
func newActor(config ActorConfig) *Actor {
	return &Actor{
		Name:               config.Name,
		StateType:          config.StateType,
		Kind:               config.Kind,
		Stateful:           config.Stateful,
		SnapshotTimeout:    config.SnapshotTimeout,
		DeactivatedTimeout: config.DeactivatedTimeout,
		Actions:            make(map[string]ActionHandler),
	}
}
