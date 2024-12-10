# Actors

his is an example of what kind of Actors you can create with Spawn

## Named Actor

In this example we are creating an actor in a Named way, that is, it is a known actor at compile time. Or a 'global' actor with only one name.

```go
actorConfig := spawn.ActorConfig{
    Name:               "UserActor",         // Nome do ator
    StateType:          &actors.UserState{}, // State type
    Kind:               spawn.Named,         // Actor Type (Named)
    Stateful:           true,                // Stateful actor
    SnapshotTimeout:    60,                  // Snapshot timeout
    DeactivatedTimeout: 120,                 // Deactivation timeout
}
```

```go
// Creates an actor with the given configuration
actor := system.BuildActor(actorConfig)

// Define uma ação simples para o ator
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
```