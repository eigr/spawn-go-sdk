package main

import (
	"log"
	"time"

	domain "examples/actors"
	logic "examples/logic"

	actorSystem "github.com/eigr/spawn-go-sdk/spawn/system"
)

func main() {
	userActor := logic.BuilUserActor()

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
