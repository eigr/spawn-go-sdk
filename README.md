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

option go_package = "github.com/eigr/spawn-go-sdk/examples/actors;actors";

message UserState {
  string name = 1;
}

message ChangeUserNamePayload {
  string new_name = 1;
}

message ChangeUserNameResponse {
  enum Status {
    OK = 0;
    ERROR = 1;
  }
  Status status = 1;
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

	"github.com/eigr/spawn-go-sdk/examples/actors"
	"github.com/eigr/spawn-go-sdk/spawn"
	"google.golang.org/protobuf/proto"
)

func main() {
	system := spawn.NewSystem()

	actorConfig := spawn.ActorConfig{
		Name:               "UserActor",
		StateType:          &actors.UserState{},
		Kind:               spawn.Named,
		Stateful:           true,
		SnapshotTimeout:    60,
		DeactivatedTimeout: 120,
	}

	actor := system.BuildActor(actorConfig)

	actor.AddAction("ChangeUserName", func(ctx spawn.ActorContext, payload proto.Message) (spawn.Value, error) {
		input, ok := payload.(*actors.ChangeUserNamePayload)
		if !ok {
			return spawn.Value{}, fmt.Errorf("invalid payload type")
		}

		state := &actors.UserState{Name: input.NewName}
		response := &actors.ChangeUserNameResponse{Status: actors.ChangeUserNameResponse_OK}

		return spawn.Of(state, response), nil
	})

	if err := system.Register(actor); err != nil {
		log.Fatalf("Failed to register actor: %v", err)
	}
}
```

## 📚 Explore More

Check out our examples folder for additional use cases and inspiration.

## 💡 Why Spawn Matters

CTOs, Tech Leads, and Developers love Spawn for its simplicity, scalability, and flexibility. Build reliable, distributed systems faster than ever.

Unleash the power of polyglot distributed systems with Spawn today!