package main

import (
	"fmt"
	"log"
	"time"

	domain "examples/actors"

	spawn "github.com/eigr/spawn-go-sdk/spawn/actors"
	actorSystem "github.com/eigr/spawn-go-sdk/spawn/system"
	"google.golang.org/protobuf/proto"
)

func main() {
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

	// Initializes the Spawn system
	system := actorSystem.NewSystem("spawn-system").
		UseProxyPort(9001).
		ExposePort(8090).
		RegisterActor(userActor)

	// Start the system
	if err := system.Start(); err != nil {
		log.Fatalf("Failed to start Actor System: %v", err)
	}

	time.Sleep(5 * time.Second)

	resp, _ := system.Invoke(
		"spawn-system",
		"UserActor",
		"ChangeUserName",
		&domain.ChangeUserNamePayload{NewName: "John Doe"},
		actorSystem.Options{})

	log.Printf("Response: %v", resp)

	system.Await()
}
