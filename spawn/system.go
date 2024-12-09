package spawn

import (
	"fmt"
)

// System represents the Spawn system.
type System struct {
	actors map[string]*Actor
}

// NewSystem creates a new Spawn system.
func NewSystem() *System {
	return &System{
		actors: make(map[string]*Actor),
	}
}

// Register registers the system in the Spawn sidecar.
func (s *System) Register(actor *Actor) error {
	// Simulação de registro
	fmt.Println("Spawn System registered")
	s.actors[actor.Name] = actor
	return nil
}

// BuildActor builds an actor in the system.
func (s *System) BuildActor(config ActorConfig) *Actor {
	actor := NewActor(config)
	return actor
}
