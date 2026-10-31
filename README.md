# protodump

Protodump is a utility to dump all Protobuf file descriptors from a given binary as *.proto files:

![Demo](https://raw.githubusercontent.com/arkadiyt/protodump/main/demo/demo.gif)

## Usage

```
go install github.com/arkadiyt/protodump/cmd/protodump@latest
./protodump -file <file to extract from> -output <output directory>
```

## Metadata preservation

The normal dump preserves map fields and descriptor-provided `json_name` values.
Maps use canonical `map<K, V>` syntax instead of synthetic entry declarations.

This preserves map semantics and JSON field names, not every protobuf option
or the original source formatting.

## Tests

```sh
go test ./...
go vet ./...
```

Fixture tests need `protoc` on PATH, without `protoc-gen-go`. They recompile
rendered schemas to check map semantics, JSON names, and optional-field metadata.

## Getting in touch

Feel free to contact me on Signal: @arkadiyt.01
