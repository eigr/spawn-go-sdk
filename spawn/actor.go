package spawn

import (
	"sync"

	"google.golang.org/protobuf/proto"
)

// ActionHandler defines the function of a Protobuf-supported action.
type ActionHandler func(ctx ActorContext, payload proto.Message) (Value, error)

// Actor represents an actor in Spawn.
type Actor struct {
	Name               string
	StateType          proto.Message
	Kind               Kind
	Stateful           bool
	SnapshotTimeout    int64
	DeactivatedTimeout int64
	actions            map[string]ActionHandler
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
}

// NewActor creates a new actor instance.
func NewActor(config ActorConfig) *Actor {
	return &Actor{
		Name:               config.Name,
		StateType:          config.StateType,
		Kind:               config.Kind,
		Stateful:           config.Stateful,
		SnapshotTimeout:    config.SnapshotTimeout,
		DeactivatedTimeout: config.DeactivatedTimeout,
		actions:            make(map[string]ActionHandler),
	}
}

// AddAction adds a new action to the actor.
func (a *Actor) AddAction(name string, handler ActionHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.actions[name] = handler
}
