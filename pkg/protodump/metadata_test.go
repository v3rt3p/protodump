package protodump

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func metadataField(name string, number int32, kind descriptorpb.FieldDescriptorProto_Type, jsonName *string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name: proto.String(name), Number: proto.Int32(number), Type: kind.Enum(),
		Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), JsonName: jsonName,
	}
}

func TestMapAndJSONMetadata(t *testing.T) {
	entry := &descriptorpb.DescriptorProto{
		Name: proto.String("ItemsEntry"), Options: &descriptorpb.MessageOptions{MapEntry: proto.Bool(true)},
		Field: []*descriptorpb.FieldDescriptorProto{
			metadataField("key", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, nil),
			metadataField("value", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING, nil),
		},
	}
	mapField := metadataField("Items", 1, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, proto.String("items_by_name"))
	mapField.Label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()
	mapField.TypeName = proto.String(".test.Container.ItemsEntry")
	ordinary := metadataField("Rows", 2, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, nil)
	ordinary.Label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()
	ordinary.TypeName = proto.String(".test.Container.Row")
	file := &descriptorpb.FileDescriptorProto{
		Name: proto.String("metadata.proto"), Package: proto.String("test"), Syntax: proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Container"), NestedType: []*descriptorpb.DescriptorProto{entry, {Name: proto.String("Row")}},
			Field: []*descriptorpb.FieldDescriptorProto{mapField, ordinary,
				metadataField("DisplayName", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING, proto.String("display_name")),
				metadataField("plain", 4, descriptorpb.FieldDescriptorProto_TYPE_STRING, nil),
				metadataField("explicit", 5, descriptorpb.FieldDescriptorProto_TYPE_STRING, proto.String("explicit")),
			},
		}},
	}
	definition, err := NewFromDescriptor(file)
	if err != nil {
		t.Fatal(err)
	}
	text := definition.String()
	for _, expected := range []string{
		`map<string, string> Items = 1 [json_name = "items_by_name"];`,
		`repeated .test.Container.Row Rows = 2;`,
		`string DisplayName = 3 [json_name = "display_name"];`,
		`string plain = 4;`,
		`string explicit = 5 [json_name = "explicit"];`,
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("missing %q in:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "message ItemsEntry") || strings.Contains(text, "repeated map<") {
		t.Errorf("synthetic map entry leaked into output:\n%s", text)
	}
}

func TestJSONNameAndDefaultOptions(t *testing.T) {
	field := metadataField("DisplayName", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, proto.String("display_name"))
	field.DefaultValue = proto.String("quoted \"value\"\\line\n")
	file := &descriptorpb.FileDescriptorProto{
		Name: proto.String("defaults.proto"), Syntax: proto.String("proto2"),
		MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Message"), Field: []*descriptorpb.FieldDescriptorProto{field}}},
	}
	definition, err := NewFromDescriptor(file)
	if err != nil {
		t.Fatal(err)
	}
	want := `optional string DisplayName = 1 [default = "quoted \"value\"\\line\n", json_name = "display_name"];`
	if !strings.Contains(definition.String(), want) {
		t.Fatalf("missing %q in:\n%s", want, definition.String())
	}
}

func TestJSONNameEscaping(t *testing.T) {
	field := metadataField("Field", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, proto.String("quoted\"slash\\name"))
	file := &descriptorpb.FileDescriptorProto{
		Name: proto.String("escaping.proto"), Syntax: proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Message"), Field: []*descriptorpb.FieldDescriptorProto{field}}},
	}
	definition, err := NewFromDescriptor(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(definition.String(), `json_name = "quoted\"slash\\name"`) {
		t.Fatal(definition.String())
	}
}
