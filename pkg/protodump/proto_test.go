package protodump

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const FIXTURES = "fixtures"

func convertProtoToFileDescriptor(filePath string) (*descriptorpb.FileDescriptorProto, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	filename := "proto.bin"
	cmd := exec.Command("protoc", "-I"+path.Dir(filePath), fmt.Sprintf("--descriptor_set_out=%s", path.Join(dir, filename)), path.Base(filePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print(string(output))
		return nil, err
	}

	data, err := os.ReadFile(path.Join(dir, filename))
	if err != nil {
		return nil, err
	}

	var descriptor descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(data, &descriptor)
	if err != nil {
		return nil, err
	}
	return descriptor.GetFile()[0], nil
}

func assertMetadataEqual(t *testing.T, before, after []*descriptorpb.DescriptorProto) {
	t.Helper()
	require.Len(t, after, len(before))
	for i, message := range before {
		other := after[i]
		assert.Equal(t, message.GetName(), other.GetName())
		assert.Equal(t, message.GetOptions().GetMapEntry(), other.GetOptions().GetMapEntry())
		require.Len(t, other.GetField(), len(message.GetField()))
		fields := make(map[int32]*descriptorpb.FieldDescriptorProto)
		for _, field := range other.GetField() {
			fields[field.GetNumber()] = field
		}
		for _, field := range message.GetField() {
			actual := fields[field.GetNumber()]
			require.NotNil(t, actual)
			assert.Equal(t, field.GetName(), actual.GetName())
			assert.Equal(t, field.GetJsonName(), actual.GetJsonName())
			assert.Equal(t, field.GetProto3Optional(), actual.GetProto3Optional())
		}
		assertMetadataEqual(t, message.GetNestedType(), other.GetNestedType())
	}
}

func TestDefinitions(t *testing.T) {
	files, err := os.ReadDir(FIXTURES)
	require.NoError(t, err)

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			filePath := path.Join(FIXTURES, file.Name())
			descriptor, err := convertProtoToFileDescriptor(filePath)
			require.NoError(t, err)

			expected, err := os.ReadFile(filePath)
			require.NoError(t, err)

			actual, err := NewFromDescriptor(descriptor)
			require.NoError(t, err)

			assert.Equal(t, string(expected), actual.String())

			// Recompile the source to check field metadata, not only formatting.
			rendered := path.Join(t.TempDir(), "rendered.proto")
			require.NoError(t, os.WriteFile(rendered, []byte(actual.String()), 0600))
			roundTrip, err := convertProtoToFileDescriptor(rendered)
			require.NoError(t, err)
			assertMetadataEqual(t, descriptor.GetMessageType(), roundTrip.GetMessageType())
		})
	}
}
