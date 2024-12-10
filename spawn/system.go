package spawn

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	protocol "spawn/eigr/functions/protocol/actors"

	"google.golang.org/protobuf/proto"
)

// System represents the Spawn system.
type System struct {
	actors     map[string]*Actor
	name       string
	proxyPort  int
	exposePort int
	url        string
}

// NewSystem creates a new Spawn system.
func NewSystem(name string) *System {
	return &System{
		actors: make(map[string]*Actor),
		name:   name,
		url:    "http://localhost", // Default URL
	}
}

// UseProxyPort sets the proxy port for the system.
func (s *System) UseProxyPort(port int) *System {
	s.proxyPort = port
	return s
}

// ExposePort sets the port to expose the ActorHost.
func (s *System) ExposePort(port int) *System {
	s.exposePort = port
	return s
}

// RegisterActor registers a single actor in the system.
func (s *System) RegisterActor(actor *Actor) *System {
	s.actors[actor.Name] = actor
	return s
}

// Start initializes the system by registering all configured actors with the sidecar.
func (s *System) Start() error {
	if len(s.actors) == 0 {
		return fmt.Errorf("no actors registered in the system")
	}

	// Converts actors into a Protobuf representation map
	actorProtos := s.convertActorsToProtobuf()

	registration := &protocol.RegistrationRequest{
		ServiceInfo: &protocol.ServiceInfo{
			ServiceName:           "spawn-go-sdk",
			ServiceVersion:        "v0.1.0",
			ServiceRuntime:        "go1.23",
			SupportLibraryName:    "spawn-go-sdk",
			SupportLibraryVersion: "v0.1.0",
			ProtocolMajorVersion:  1,
			ProtocolMinorVersion:  1,
		},
		ActorSystem: &protocol.ActorSystem{
			Name: s.name,
			Registry: &protocol.Registry{
				Actors: actorProtos,
			},
		},
	}

	data, err := proto.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to serialize registration request: %w", err)
	}

	resp, err := s.postToSidecar(data)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register actors, status code: %d", resp.StatusCode)
	}

	log.Println("Actors successfully registered and system started")
	return nil
}

// postToSidecar sends the serialized data to the Spawn sidecar API.
func (s *System) postToSidecar(data []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s:%d/api/v1/system", s.url, s.proxyPort)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "user-function-client/0.1.0")
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	return client.Do(req)
}

// convertActorsToProtobuf converts the registered actors into a map with actor names as keys and their Protobuf representation as values.
func (s *System) convertActorsToProtobuf() map[string]*protocol.Actor {
	actorMap := make(map[string]*protocol.Actor, len(s.actors))

	for _, actor := range s.actors {
		// Converting actions
		actions := make([]*protocol.Action, 0, len(actor.actions))
		for actionName := range actor.actions {
			actions = append(actions, &protocol.Action{
				Name: actionName,
			})
		}

		// Creating snapshot strategy
		var snapshotStrategy *protocol.ActorSnapshotStrategy
		if actor.SnapshotTimeout > 0 {
			snapshotStrategy = &protocol.ActorSnapshotStrategy{
				Strategy: &protocol.ActorSnapshotStrategy_Timeout{
					Timeout: &protocol.TimeoutStrategy{
						Timeout: actor.SnapshotTimeout,
					},
				},
			}
		}

		// Creating deactivation strategy
		var deactivationStrategy *protocol.ActorDeactivationStrategy
		if actor.DeactivatedTimeout > 0 {
			deactivationStrategy = &protocol.ActorDeactivationStrategy{
				Strategy: &protocol.ActorDeactivationStrategy_Timeout{
					Timeout: &protocol.TimeoutStrategy{
						Timeout: actor.DeactivatedTimeout,
					},
				},
			}
		}

		// Configuring ActorSettings
		settings := &protocol.ActorSettings{
			Kind:                 mapKindFromGoToProto(actor.Kind),
			Stateful:             actor.Stateful,
			SnapshotStrategy:     snapshotStrategy,
			DeactivationStrategy: deactivationStrategy,
		}

		// Configuring pool size if the actor kind is pooled
		if actor.Kind == Pooled {
			settings.MinPoolSize = actor.MinPoolSize
			settings.MaxPoolSize = actor.MaxPoolSize
		}

		// Adding the actor to the map
		actorMap[actor.Name] = &protocol.Actor{
			Id: &protocol.ActorId{
				Name:   actor.Name,
				System: s.name,
			},
			State:        &protocol.ActorState{},
			Metadata:     &protocol.Metadata{},
			Settings:     settings,
			Actions:      actions,
			TimerActions: nil, // TODO: Implement timer actions
		}
	}

	return actorMap
}

// BuildActor creates an actor and returns it.
func (s *System) BuildActor(config ActorConfig) *Actor {
	return ActorOf(config)
}

func mapKindFromGoToProto(kind Kind) protocol.Kind {
	switch kind {
	case Named:
		return protocol.Kind_NAMED
	case Unnamed:
		return protocol.Kind_UNNAMED
	case Pooled:
		return protocol.Kind_POOLED
	case Task:
		return protocol.Kind_TASK
	case Projection:
		return protocol.Kind_PROJECTION
	default:
		return protocol.Kind_UNKNOW_KIND
	}
}
