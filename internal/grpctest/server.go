// Package grpctest provides a minimal in-process gRPC server, built purely
// from hand-constructed descriptors (no protoc/generated code required), so
// pkg/grpcclient and pkg/handlers tests can exercise apix's reflection-based
// method resolution and invocation end-to-end.
package grpctest

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Names of the single service/method this fake server exposes.
const (
	ServiceName = "testpkg.TestService"
	MethodName  = "GetReply"
)

// Server is a real, listening gRPC server exposing exactly one unary
// method: testpkg.TestService/GetReply(Empty) returns (Reply), where Reply
// has a single string field "message".
type Server struct {
	Listener net.Listener

	grpcSrv *grpc.Server

	// lastMetadata captures the incoming metadata of the most recently
	// handled call, for tests asserting on metadata/header propagation.
	lastMetadata metadata.MD
}

// NewServer builds the server and starts it listening on a local ephemeral
// port, but does not yet accept connections - call Serve to do that.
func NewServer() (*Server, error) {
	fileDesc, replyDesc, emptyDesc, err := buildDescriptors()
	if err != nil {
		return nil, fmt.Errorf("failed to build test descriptors: %w", err)
	}

	filesReg := &protoregistry.Files{}
	if err := filesReg.RegisterFile(fileDesc); err != nil {
		return nil, fmt.Errorf("failed to register test descriptors: %w", err)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	s := &Server{Listener: lis}
	s.grpcSrv = grpc.NewServer()

	s.grpcSrv.RegisterService(&grpc.ServiceDesc{
		ServiceName: ServiceName,
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: MethodName,
				Handler: func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
					in := dynamicpb.NewMessage(emptyDesc)
					if err := dec(in); err != nil {
						return nil, err
					}

					if md, ok := metadata.FromIncomingContext(ctx); ok {
						s.lastMetadata = md
					}

					reply := dynamicpb.NewMessage(replyDesc)
					reply.Set(replyDesc.Fields().ByName("message"), protoreflect.ValueOfString("hello"))
					return reply, nil
				},
			},
		},
		Metadata: "testservice.proto",
	}, nil)

	reflectionSrv := reflection.NewServerV1(reflection.ServerOptions{
		Services:           s.grpcSrv,
		DescriptorResolver: filesReg,
	})
	reflectionpb.RegisterServerReflectionServer(s.grpcSrv, reflectionSrv)

	return s, nil
}

// Serve blocks accepting connections; run it in a goroutine.
func (s *Server) Serve() error {
	return s.grpcSrv.Serve(s.Listener)
}

// Addr returns the "host:port" the server is listening on, suitable for use
// as a config.Domain.Base.
func (s *Server) Addr() string {
	return s.Listener.Addr().String()
}

// LastMetadata returns the incoming gRPC metadata captured on the most
// recent handled call, or nil if no call has been handled yet.
func (s *Server) LastMetadata() metadata.MD {
	return s.lastMetadata
}

// Close stops the server and releases its listener.
func (s *Server) Close() {
	s.grpcSrv.Stop()
}

func buildDescriptors() (file protoreflect.FileDescriptor, reply, empty protoreflect.MessageDescriptor, err error) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("testservice.proto"),
		Package: proto.String("testpkg"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Empty")},
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
				Name: proto.String("TestService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String(MethodName),
						InputType:  proto.String(".testpkg.Empty"),
						OutputType: proto.String(".testpkg.Reply"),
					},
				},
			},
		},
	}

	fileDesc, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
	if err != nil {
		return nil, nil, nil, err
	}

	return fileDesc, fileDesc.Messages().Get(1), fileDesc.Messages().Get(0), nil
}
