package system

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eigr/spawn-go-sdk/spawn/actors"
	protocol "github.com/eigr/spawn-go-sdk/spawn/eigr/functions/protocol/actors"

	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

// System represents the Spawn system.
type System struct {
	actors     map[string]*actors.Actor
	name       string
	proxyPort  int
	exposePort int
	url        string
	stopCh     chan struct{}
	server     *http.Server
	wg         sync.WaitGroup
}

type invocationOptions map[string]interface{}

// Creating an alias for InvocationOptions, now called Options
type Options = invocationOptions

// NewSystem creates a new Spawn system.
func NewSystem(name string) *System {
	return &System{
		actors: make(map[string]*actors.Actor),
		name:   name,
		url:    "http://localhost", // Default URL
		stopCh: make(chan struct{}),
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
func (s *System) RegisterActor(actor *actors.Actor) *System {
	s.actors[actor.Name] = actor
	return s
}

// BuildActor creates an actor and returns it.
func (s *System) BuildActor(config actors.ActorConfig) *actors.Actor {
	return actors.ActorOf(config)
}

// Start initializes the system by registering all configured actors with the sidecar.
func (s *System) Start() error {
	if len(s.actors) == 0 {
		return fmt.Errorf("no actors registered in the system")
	}

	go s.startServer()

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

	go s.listenForTermination()

	log.Println("Actors successfully registered and system started")
	///s.wg.Wait()
	return nil
}

// Await waits for the system to stop.
func (s *System) Await() {
	s.wg.Add(1)
	defer s.wg.Done()
	// Wait until `stopCh` channel is closed
	<-s.stopCh
}

// client API
func (s *System) Invoke(system string, actorName string, action string, request proto.Message, options Options) (proto.Message, error) {
	log.Printf("Invoking actor: %s, action: %s", actorName, action)
	parent, hasParent := options["parent"]
	async, hasAsync := options["async"]
	metadata, hasMetadata := options["metadata"]

	req := &protocol.InvocationRequest{}
	req.System = &protocol.ActorSystem{Name: system}

	actor := &protocol.Actor{
		Id: &protocol.ActorId{
			Name:   actorName,
			System: system,
		},
	}

	if hasParent {
		if parentStr, ok := parent.(string); ok {
			actor.Id.Parent = parentStr
			req.RegisterRef = actorName
		} else {
			return nil, fmt.Errorf("parent must be a string")
		}
	}
	if hasAsync {
		if asyncBool, ok := async.(bool); ok {
			req.Async = asyncBool
		} else {
			return nil, fmt.Errorf("async must be a bool")
		}
	}
	if hasMetadata {
		req.Metadata = metadata.(map[string]string)
	}

	req.Actor = actor
	req.ActionName = action

	if request != nil {
		payload, err := anypb.New(request)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload to make an Actor InvocationRequest: %w", err)
		}

		req.Payload = &protocol.InvocationRequest_Value{Value: payload}
	}

	r, err := proto.Marshal(req)
	if err != nil {
		log.Fatalln("Failed to encode Actor InvocationRequest:", err)
	}

	// call proxy to invoke actor
	responseBytes, err := s.invokeActor(actorName, r)
	if err != nil {
		return nil, err
	}

	resp := &protocol.InvocationResponse{}
	if err := proto.Unmarshal(responseBytes, resp); err != nil {
		log.Fatalln("Failed to parse Actor InvocationResponse:", err)
	}

	log.Printf("Actor invocation response: %v", resp)

	if resp.Status.GetStatus() != protocol.Status_OK {
		return nil, fmt.Errorf("actor invocation failed: %s", resp.Status)
	}

	var message proto.Message

	switch p := resp.GetPayload().(type) {
	case *protocol.InvocationResponse_Value:
		iany := p.Value
		log.Printf("Received response payload: %v", iany)
		msg, err := unmarshalAny(iany)
		log.Printf("Unmarshalled response payload: %v", msg)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling response value: %v", err)
		}
		message = msg
	case *protocol.InvocationResponse_Noop:
		message = nil
	default:
		return nil, fmt.Errorf("unknown payload type")
	}

	return message, nil
}

// private functions

func (s *System) listenForTermination() {
	// Add a goroutine to the WaitGroup to wait for its completion
	s.wg.Add(1)
	defer s.wg.Done()

	// Create a channel to capture signals
	signalChan := make(chan os.Signal, 1)

	// Report SIGINT and SIGTERM signals
	signal.Notify(signalChan, syscall.SIGTERM)

	// Block until a termination signal is received
	sig := <-signalChan
	log.Printf("Received %s, shutting down gracefully...", sig)

	// Tenta fechar o servidor HTTP
	if err := s.server.Close(); err != nil {
		log.Printf("Error closing the server: %v", err)
	}

	close(s.stopCh)
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
		actions := make([]*protocol.Action, 0, len(actor.Actions))
		for actionName := range actor.Actions {
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
		if actor.Kind == actors.Pooled {
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

func mapKindFromGoToProto(kind actors.Kind) protocol.Kind {
	switch kind {
	case actors.Named:
		return protocol.Kind_NAMED
	case actors.Unnamed:
		return protocol.Kind_UNNAMED
	case actors.Pooled:
		return protocol.Kind_POOLED
	case actors.Task:
		return protocol.Kind_TASK
	case actors.Projection:
		return protocol.Kind_PROJECTION
	default:
		return protocol.Kind_UNKNOW_KIND
	}
}

func (s *System) startServer() {
	http.HandleFunc("/api/v1/actors/actions", s.handleActorInvocation)
	addr := fmt.Sprintf(":%d", s.exposePort)
	log.Printf("ActorHost server started on port %d\n", s.exposePort)

	s.server = &http.Server{Addr: addr}

	// Adds the goroutine to the WaitGroup to wait for its completion
	s.wg.Add(1)
	defer s.wg.Done()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ActorHost server failed: %v", err)
	}
}

func (s *System) handleActorInvocation(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusInternalServerError)
		return
	}

	var actorInvocation protocol.ActorInvocation
	if err := proto.Unmarshal(body, &actorInvocation); err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal protobuf: %v", err), http.StatusBadRequest)
		return
	}

	// Process the invocation
	log.Printf("Received actor invocation: %v", &actorInvocation)
	resp := s.processActorInvocation(&actorInvocation)

	payloadBytes, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal protobuf response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(payloadBytes)
}

func (s *System) processActorInvocation(actorInvocation *protocol.ActorInvocation) *protocol.ActorInvocationResponse {
	log.Printf("Processing actor invocation: %v", actorInvocation)

	actorName := actorInvocation.Actor.Name
	actionName := actorInvocation.ActionName
	requestContext := actorInvocation.CurrentContext
	actualStateAny := requestContext.State

	var req proto.Message
	switch payload := actorInvocation.Payload.(type) {
	case *protocol.ActorInvocation_Value:
		// Deserialize the payload
		request, err := unmarshalAny(payload.Value)
		if err != nil {
			log.Printf("Failed to unmarshal payload: %v", err)
			return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
		}
		req = request
	case *protocol.ActorInvocation_Noop:
		log.Printf("No operation payload received for actor: %s", actorName)
	}

	actor, ok := s.actors[actorName]
	if !ok {
		log.Printf("Actor not found: %s", actorName)
		return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
	}

	actionHandler, ok := actor.Actions[actionName]
	if !ok {
		log.Printf("Action not found: %s for actor %s", actionName, actorName)
		return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
	}

	fmt.Printf("Marshalled Any: %v\n", actualStateAny)
	// Unmarshal the actor's current state
	var stateValue proto.Message
	if actualStateAny == nil {
		log.Printf("State not found for actor %s", actorName)
	} else {
		actualStateValue, err := unmarshalAny(actualStateAny)
		if err != nil {
			log.Printf("Failed to unmarshal state for actor %s: %v", actorName, err)
			return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
		}

		stateValue = actualStateValue
	}

	// Invoke the action handler
	value, err := actionHandler(&actors.ActorContext{CurrentState: stateValue}, req)
	if err != nil {
		log.Printf("Error invoking action: %s for actor %s, error: %v", actionName, actorName, err)
		return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
	}

	log.Printf("Action [%s] response: %s for actor %s", actionName, value, actorName)

	// Marshal the returned value into an Any type
	//payloadAny, err := anypb.New(value)
	// if err != nil {
	// 	log.Printf("Failed to marshal response payload: %v", err)
	// 	return &protocol.ActorInvocationResponse{}
	// }

	// Create the updated context
	var updatedState *anypb.Any = actualStateAny
	if value.State != nil {
		us, err := anypb.New(value.State)
		if err != nil {
			log.Printf("Failed to marshal updated state: %v", err)
			return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
		}

		updatedState = us
	}

	updatedContext := &protocol.Context{
		State: updatedState,
	}

	log.Printf("Value Response: %v", value.Response)
	responPayload, err := anypb.New(value.Response)
	if err != nil {
		log.Printf("Failed to marshal response payload: %v", err)
		return &protocol.ActorInvocationResponse{ActorName: actorName, ActorSystem: s.name}
	}

	return &protocol.ActorInvocationResponse{
		ActorName:      actorName,
		ActorSystem:    s.name,
		UpdatedContext: updatedContext,
		Payload:        &protocol.ActorInvocationResponse_Value{Value: responPayload},
		Workflow:       nil,   // Populate if needed
		Checkpoint:     false, // Example: enable checkpointing
	}
}

func unmarshalAny(iany *anypb.Any) (proto.Message, error) {
	if iany == nil {
		return nil, fmt.Errorf("input Any message is nil")
	}

	// Extract the message type name from the TypeUrl
	msgName := strings.TrimPrefix(iany.GetTypeUrl(), "type.googleapis.com/")

	// Lookup the message type in the global registry
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgName))
	if err != nil {
		return nil, fmt.Errorf("message type %s not found: %v", msgName, err)
	}

	// Create a new instance of the message type
	message := mt.New().Interface()

	// Unmarshal the Any value into the message instance
	err = proto.Unmarshal(iany.GetValue(), message)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling failed: %v", err)
	}

	return message, nil
}

func (s *System) invokeActor(actorName string, requestBytes []byte) ([]byte, error) {
	// Monta a URL de invocação do ator remoto
	url := fmt.Sprintf("%s:%d/api/v1/system/%s/actors/%s/invoke", s.url, s.proxyPort, s.name, actorName)

	// Configura os cabeçalhos HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("User-Agent", "user-function-client/0.1.0")
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Content-Type", "application/octet-stream")

	// Envia a requisição HTTP POST
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verifica a resposta
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("falha na invocação do ator. Status: %d, Erro: %s", resp.StatusCode, string(body))
	}

	// Lê o conteúdo da resposta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	return respBody, nil
}
