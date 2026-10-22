package grpcclient_test

import (
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/grpcclient"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func buildFormatTestDescriptors(t *testing.T) (requestDesc, replyDesc protoreflect.MessageDescriptor) {
	t.Helper()

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("format_test.proto"),
		Package: proto.String("formattest"),
		Syntax:  proto.String("proto3"),
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("Status"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("UNKNOWN"), Number: proto.Int32(0)},
					{Name: proto.String("ACTIVE"), Number: proto.Int32(1)},
					{Name: proto.String("INACTIVE"), Number: proto.Int32(2)},
				},
			},
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Address"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("country"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: proto.String("country"),
					},
					{
						Name:     proto.String("previous"),
						Number:   proto.Int32(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".formattest.Address"),
						JsonName: proto.String("previous"),
					},
				},
			},
			{
				Name: proto.String("Request"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("id"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: proto.String("id"),
					},
					{
						Name:     proto.String("tags"),
						Number:   proto.Int32(2),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: proto.String("tags"),
					},
					{
						Name:     proto.String("labels"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".formattest.Request.LabelsEntry"),
						JsonName: proto.String("labels"),
					},
					{
						Name:     proto.String("address"),
						Number:   proto.Int32(4),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".formattest.Address"),
						JsonName: proto.String("address"),
					},
					{
						Name:     proto.String("status"),
						Number:   proto.Int32(5),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: proto.String(".formattest.Status"),
						JsonName: proto.String("status"),
					},
					{
						Name:       proto.String("email"),
						Number:     proto.Int32(6),
						Label:      descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName:   proto.String("email"),
						OneofIndex: proto.Int32(0),
					},
					{
						Name:       proto.String("phone"),
						Number:     proto.Int32(7),
						Label:      descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:       descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName:   proto.String("phone"),
						OneofIndex: proto.Int32(0),
					},
					{
						Name:           proto.String("nickname"),
						Number:         proto.Int32(8),
						Label:          descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:           descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName:       proto.String("nickname"),
						OneofIndex:     proto.Int32(1),
						Proto3Optional: proto.Bool(true),
					},
				},
				NestedType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("LabelsEntry"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{Name: proto.String("key"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("key")},
							{Name: proto.String("value"), Number: proto.Int32(2), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("value")},
						},
						Options: &descriptorpb.MessageOptions{MapEntry: proto.Bool(true)},
					},
				},
				OneofDecl: []*descriptorpb.OneofDescriptorProto{
					{Name: proto.String("contact")},
					{Name: proto.String("_nickname")},
				},
			},
			{
				Name: proto.String("Reply"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("message"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: proto.String("message"),
					},
				},
			},
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("FormatTestService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String("DoThing"),
						InputType:  proto.String(".formattest.Request"),
						OutputType: proto.String(".formattest.Reply"),
					},
				},
			},
		},
	}

	fileDesc, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatalf("failed to build test descriptors: %v", err)
	}

	// Messages, in declaration order: Address(0), Request(1), Reply(2).
	return fileDesc.Messages().Get(1), fileDesc.Messages().Get(2)
}

func TestFormatMethodSummary(t *testing.T) {
	requestDesc, replyDesc := buildFormatTestDescriptors(t)

	info := &grpcclient.MethodInfo{
		FullName: "/formattest.FormatTestService/DoThing",
		Input:    requestDesc,
		Output:   replyDesc,
	}

	out := grpcclient.FormatMethodSummary(info)

	wantContains := []string{
		"formattest.FormatTestService/DoThing",
		"Request: formattest.Request",
		"Response: formattest.Reply",
		"id string (singular, #1)",
		"tags string (repeated, #2)",
		"labels map<string, string> (map, #3)",
		"address formattest.Address (singular, #4)",
		"country string (singular, #1)",
		"status formattest.Status (singular, #5)",
		"values: UNKNOWN, ACTIVE, INACTIVE",
		"email string (singular, oneof contact, #6)",
		"phone string (singular, oneof contact, #7)",
		"nickname string (singular, explicit presence, #8)",
		"message string (singular, #1)",
	}

	for _, want := range wantContains {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestFormatMethodSummary_GuardsAgainstRecursion(t *testing.T) {
	requestDesc, replyDesc := buildFormatTestDescriptors(t)

	info := &grpcclient.MethodInfo{
		FullName: "/formattest.FormatTestService/DoThing",
		Input:    requestDesc,
		Output:   replyDesc,
	}

	out := grpcclient.FormatMethodSummary(info)

	if !strings.Contains(out, "previous formattest.Address (repeated, #2)") {
		t.Errorf("expected the self-referential field to be listed, got:\n%s", out)
	}

	if !strings.Contains(out, "recursive reference to formattest.Address") {
		t.Errorf("expected recursion into Address.previous to be guarded, got:\n%s", out)
	}
}

func TestFormatMethodSummary_Streaming(t *testing.T) {
	requestDesc, replyDesc := buildFormatTestDescriptors(t)

	info := &grpcclient.MethodInfo{
		FullName:          "/formattest.FormatTestService/DoThing",
		Input:             requestDesc,
		Output:            replyDesc,
		IsClientStreaming: true,
	}

	out := grpcclient.FormatMethodSummary(info)

	if !strings.Contains(out, "Streaming: client=true, server=false") {
		t.Errorf("expected streaming flags to be printed, got:\n%s", out)
	}
}
