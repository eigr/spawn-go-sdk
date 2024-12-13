package logic

import (
	"fmt"
	"log"

	domain "examples/actors"

	spawn "github.com/eigr/spawn-go-sdk/spawn/actors"

	"google.golang.org/protobuf/proto"
)

func BuilUserActor() *spawn.Actor {
	// Defines the actor configuration
	actorConfig := spawn.ActorConfig{
		Name:               "UserActor",         // Name of ator
		StateType:          &domain.UserState{}, // State type
		Kind:               spawn.Named,         // Actor Type (Named)
		Stateful:           true,                // Stateful actor
		SnapshotTimeout:    60,                  // Snapshot timeout
		DeactivatedTimeout: 120,                 // Deactivation timeout
	}

	// Creates an actor directly
	userActor := spawn.ActorOf(actorConfig)

	// Define a simple action for the actor
	userActor.AddAction("ChangeUserName", func(ctx *spawn.ActorContext, payload proto.Message) (spawn.Value, error) {
		// Convert payload to expected type
		log.Printf("Received invoke on Action ChangeUserName. Payload: %v", payload)
		input, ok := payload.(*domain.ChangeUserNamePayload)
		if !ok {
			return spawn.Value{}, fmt.Errorf("invalid payload type")
		}

		// Updates the status and prepares the response
		newState := &domain.UserState{Name: input.NewName}
		response := &domain.ChangeUserNameResponse{ResponseStatus: domain.ChangeUserNameResponse_OK}

		// Returns response to caller and persist new state
		return spawn.Of(response).
			State(newState).
			Materialize(), nil
	})

	return userActor
}
