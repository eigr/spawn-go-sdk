# [Spawn Go SDK](https://github.com/eigr/spawn)

**Seamless Actor Mesh Runtime for Go Developers**

---

## 🚀 **Why Choose Spawn?**

- **Developer Friendly**: Simplify building distributed, stateful applications.
- **Scalable**: Designed for cloud-native environments with polyglot support.
- **Effortless Integration**: Build robust systems with minimal boilerplate.

---

## **🌟 Features**

- Fully managed actor lifecycle.
- State persistence and snapshots.
- Polyglot SDKs for ultimate flexibility. In this case GO SDK \0/.
- Optimized for high performance and low latency.

---

## **📦 Installation**

Set up your environment in seconds. Install the Spawn CLI:

```bash
curl -sSL https://github.com/eigr/spawn/releases/download/v1.4.3/install.sh | sh
```

## 🔥 Getting Started

### 1️⃣ Create a New Project

```bash
spawn new go hello_world
```

### 2️⃣ Define Your Protocol

Leverage the power of Protobuf to define your actor's schema:

```proto
syntax = "proto3";

package examples.actors;

option go_package = "examples/actors";

message UserState {
  string name = 1;
}

message ChangeUserNamePayload {
  string new_name = 1;
}

message ChangeUserNameResponse {
  // this is a bad example, but it's just an example
  enum ResponseStatus {
    OK = 0;
    ERROR = 1;
  }
  ResponseStatus response_status = 1;
}

service UserActor {
  rpc ChangeUserName(ChangeUserNamePayload) returns (ChangeUserNameResponse) {}
}
```

### 3️⃣ Compile Your Protobuf

Follow the example in our [Makefile](./Makefile).

### 4️⃣ Implement Your Business Logic

Start writing actors with ease:

```go
package main

import (
	"fmt"
	"log"

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
	userActor.AddAction("ChangeUserName", func(ctx spawn.ActorContext, payload proto.Message) (spawn.Value, error) {
		// Convert payload to expected type
		input, ok := payload.(*actors.ChangeUserNamePayload)
		if !ok {
			return spawn.Value{}, fmt.Errorf("invalid payload type")
		}

		// Updates the status and prepares the response
		state := &actors.UserState{Name: input.NewName}
		response := &actors.ChangeUserNameResponse{ResponseStatus: actors.ChangeUserNameResponse_OK}

		// Returns status and response
		return spawn.Of(state, response), nil
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
}
```

### 5️⃣ Build and generate application in dev mode

```bash
go build -v -gcflags="all=-N -l" -o ./bin/examples/app ./examples
```

### 6️⃣ Start Spawn CLI in dev mode

```bash
spawn dev run
```

Expected output similar to this:

```bash
🏃  Starting Spawn Proxy in dev mode...
❕  Spawn Proxy uses the following mapped ports: [
      Proxy HTTP: nil:9001,
      Proxy gRPC: nil:9980
    ]
🚀  Spawn Proxy started successfully in dev mode. Container Id: 7961ed391e06b36c6d73290a6d6a0cbb60cf910fd4e4ff3fc5b7bd49ed677976
```

### 7️⃣ Start application in dev mode

```bash
./bin/examples/app
```

### 8️⃣ Validate Spawn CLI docker logs

Use Id for previous `spawn dev run` command to see docker logs.

```bash
docker logs -f 7961ed391e06b36c6d73290a6d6a0cbb60cf910fd4e4ff3fc5b7bd49ed677976
```

Expected output similar to (ommited some logs for brevity):

```bash
2024-12-10 21:21:32.028 [proxy@GF-595]:[pid=<0.2770.0> ]:[info]:[Proxy.Config] Loading configs
2024-12-10 21:21:32.029 [proxy@GF-595]:[pid=<0.2770.0> ]:[info]:Loading config: [actor_system_name]:[my-system]
2024-12-10 21:21:32.044 [proxy@GF-595]:[pid=<0.2775.0> ]:[info]:[SUPERVISOR] Sidecar.Supervisor is up
2024-12-10 21:21:32.044 [proxy@GF-595]:[pid=<0.2777.0> ]:[info]:[SUPERVISOR] Sidecar.ProcessSupervisor is up
2024-12-10 21:21:32.045 [proxy@GF-595]:[pid=<0.2779.0> ]:[info]:[SUPERVISOR] Sidecar.MetricsSupervisor is up
2024-12-10 21:21:32.046 [proxy@GF-595]:[pid=<0.2783.0> ]:[info]:[SUPERVISOR] Spawn.Supervisor is up
2024-12-10 21:21:32.048 [proxy@GF-595]:[pid=<0.2791.0> ]:[info]:[SUPERVISOR] Spawn.Cluster.StateHandoff.ManagerSupervisor is up
2024-12-10 21:21:32.053 [proxy@GF-595]:[pid=<0.2809.0> ]:[info]:[mnesiac:proxy@GF-595] mnesiac starting, with []
2024-12-10 21:21:32.078 [proxy@GF-595]:[pid=<0.2809.0> ]:[info]:Elixir.Statestores.Adapters.Native.SnapshotStore Initialized with result {:aborted, {:already_exists, Statestores.Adapters.Native.SnapshotStore}}
2024-12-10 21:21:32.079 [proxy@GF-595]:[pid=<0.2809.0> ]:[info]:[mnesiac:proxy@GF-595] mnesiac started
2024-12-10 21:21:32.083 [proxy@GF-595]:[pid=<0.2855.0> ]:[info]:[SUPERVISOR] Actors.Supervisors.ActorSupervisor is up
2024-12-10 21:21:32.123 [proxy@GF-595]:[pid=<0.2772.0> ]:[info]:Running Proxy.Router with Bandit 1.5.2 at 0.0.0.0:9001 (http)
2024-12-10 21:21:32.124 [proxy@GF-595]:[pid=<0.2770.0> ]:[info]:Proxy Application started successfully in 0.095587ms. Running with 8 schedulers.
2024-12-10 21:21:56.518 [proxy@GF-595]:[pid=<0.3419.0> ]:[info]:POST /api/v1/system
2024-12-10 21:21:56.523 [proxy@GF-595]:[pid=<0.3419.0> ]:[info]:Sent 200 in 4ms
2024-12-10 21:21:56.526 [proxy@GF-595]:[pid=<0.3433.0> ]:[notice]:Activating Actor "UserActor" with Parent "" in Node :"proxy@GF-595". Persistence true.
2024-12-10 21:21:56.528 [proxy@GF-595]:[pid=<0.3424.0> ]:[info]:Actor UserActor Activated on Node :"proxy@GF-595" in 3402ms
```

## 📚 Explore More

Check out our examples folder for additional use cases and inspiration.

## 💡 Why Spawn Matters

CTOs, Tech Leads, and Developers love Spawn for its simplicity, scalability, and flexibility. Build reliable, distributed systems faster than ever.

Unleash the power of polyglot distributed systems with Spawn today!