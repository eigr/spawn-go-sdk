package main

import (
	"fmt"
	"log"

	"github.com/eigr/spawn-go-sdk/examples/actors"
	"github.com/eigr/spawn-go-sdk/spawn"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Initializes the Spawn system
	system := spawn.NewSystem()

	// Defines the actor configuration
	actorConfig := spawn.ActorConfig{
		Name:               "UserActor",         // Nome do ator
		StateType:          &actors.UserState{}, // State type
		Kind:               spawn.Named,         // Actor Type (Named)
		Stateful:           true,                // Stateful actor
		SnapshotTimeout:    60,                  // Snapshot timeout
		DeactivatedTimeout: 120,                 // Deactivation timeout
	}

	// Creates an actor with the given configuration
	actor := system.BuildActor(actorConfig)

	// Define a simple action for the actor
	actor.AddAction("ChangeUserName", func(ctx spawn.ActorContext, payload proto.Message) (spawn.Value, error) {
		// Convert payload to expected type
		input, ok := payload.(*actors.ChangeUserNamePayload)
		if !ok {
			return spawn.Value{}, fmt.Errorf("invalid payload type")
		}

		// Updates the status and prepares the response
		state := &actors.UserState{Name: input.NewName}
		response := &actors.ChangeUserNameResponse{Status: actors.ChangeUserNameResponse_OK}

		// Returns status and response
		return spawn.Of(state, response), nil
	})

	// Registra o ator no sistema Spawn
	if err := system.Register(actor); err != nil {
		log.Fatalf("Erro ao registrar ator: %v", err)
	}
}
