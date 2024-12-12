package main

import (
	"fmt"
	"log"
	"time"

	"examples/actors"

	"github.com/eigr/spawn-go-sdk/spawn"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Defines the actor configuration
	actorConfig := spawn.ActorConfig{
		Name:               "UserActor",         // Name of ator
		StateType:          &actors.UserState{}, // State type
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
		input, ok := payload.(*actors.ChangeUserNamePayload)
		if !ok {
			return spawn.Value{}, fmt.Errorf("invalid payload type")
		}

		// Updates the status and prepares the response
		newState := &actors.UserState{Name: input.NewName}
		response := &actors.ChangeUserNameResponse{ResponseStatus: actors.ChangeUserNameResponse_OK}

		// Returns response to caller and persist new state
		return spawn.Of(response).
			State(newState).
			Materialize(), nil
	})

	// Initializes the Spawn system
	system := spawn.NewSystem("spawn-system").
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
		&actors.ChangeUserNamePayload{NewName: "John Doe"},
		spawn.Options{})

	log.Printf("Response: %v", resp)

	system.Await()
}
