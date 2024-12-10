module github.com/eigr/spawn-go-sdk

go 1.23

require (
    google.golang.org/protobuf v1.35.2
    github.com/eigr/spawn-go-sdk/spawn v0.1.0
)

replace github.com/eigr/spawn-go-sdk/spawn => ./spawn
replace spawn/eigr/functions/protocol/actors => ./spawn/eigr/functions/protocol/actors
