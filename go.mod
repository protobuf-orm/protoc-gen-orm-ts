module github.com/protobuf-orm/protoc-gen-orm-ts

go 1.24.5

require (
	github.com/ettle/strcase v0.2.0
	github.com/protobuf-orm/protobuf-orm v0.0.0-20250823135318-4d3deece5f2e
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/lesomnus/proto-merge v0.0.0-20250820105615-abf3d11f6ba6 // indirect
	github.com/protobuf-orm/protoc-gen-orm-service v0.0.0-20250823141140-874ad79b6aea // indirect
)

tool (
	github.com/lesomnus/proto-merge
	github.com/protobuf-orm/protoc-gen-orm-service
)
