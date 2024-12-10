module github.com/eigr/spawn-go-sdk/examples

go 1.23

toolchain go1.23.0

require (
	github.com/eigr/spawn-go-sdk/spawn v0.1.0
	google.golang.org/protobuf v1.35.2
)

//replace github.com/eigr/spawn-go-sdk/spawn => ../spawn
