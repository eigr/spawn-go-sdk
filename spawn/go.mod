module "github.com/eigr/spawn-go-sdk"

go 1.23

toolchain go1.23.0

require (
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576
	google.golang.org/protobuf v1.35.2
)

replace spawn/eigr/functions/protocol/actors => ./eigr/functions/protocol/actors
