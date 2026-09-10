package grpctest

import (
	"context"
	"fmt"
	"net"
	"strconv"

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

const (
	ServiceName = "testpkg.TestService"
	MethodName  = "GetReply"
)

type Server struct {
	Listener net.Listener

	grpcSrv *grpc.Server

	lastMetadata metadata.MD
	lastName     string
}

func NewServer() (*Server, error) {
	fileDesc, replyDesc, requestDesc, err := buildDescriptors()
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
	s.grpcSrv = grpc.NewServer(grpc.StreamInterceptor(s.streamInterceptor))

	s.grpcSrv.RegisterService(&grpc.ServiceDesc{
		ServiceName: ServiceName,
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: MethodName,
				Handler: func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
					in := dynamicpb.NewMessage(requestDesc)
					if err := dec(in); err != nil {
						return nil, err
					}

					if md, ok := metadata.FromIncomingContext(ctx); ok {
						s.lastMetadata = md
					}

					name := in.Get(requestDesc.Fields().ByName("name")).String()
					s.lastName = name

					reply := dynamicpb.NewMessage(replyDesc)
					reply.Set(replyDesc.Fields().ByName("message"), protoreflect.ValueOfString("hello, "+name))
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

func (s *Server) Serve() error {
	return s.grpcSrv.Serve(s.Listener)
}

func (s *Server) Addr() string {
	return s.Listener.Addr().String()
}

// Host is suitable for a config.Domain.Base, which must never include a
// port itself.
func (s *Server) Host() string {
	host, _, err := net.SplitHostPort(s.Addr())
	if err != nil {
		return s.Addr()
	}

	return host
}

func (s *Server) Port() int {
	_, portStr, err := net.SplitHostPort(s.Addr())
	if err != nil {
		return 0
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}

	return port
}

func (s *Server) LastMetadata() metadata.MD {
	return s.lastMetadata
}

// LastName returns the "name" field of the most recent GetReply request
// body, for tests to assert that a piped JSON request body actually
// reached the server.
func (s *Server) LastName() string {
	return s.lastName
}

// streamInterceptor exists so tests can assert that metadata reaches
// streaming calls (e.g. reflection's ServerReflectionInfo), not just unary
// ones.
func (s *Server) streamInterceptor(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
		s.lastMetadata = md
	}

	return handler(srv, ss)
}

func (s *Server) Close() {
	s.grpcSrv.Stop()
}

func buildDescriptors() (file protoreflect.FileDescriptor, reply, request protoreflect.MessageDescriptor, err error) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("testservice.proto"),
		Package: proto.String("testpkg"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("GetReplyRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("name"),
						Number:   proto.Int32(1),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						JsonName: proto.String("name"),
					},
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
				Name: proto.String("TestService"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       proto.String(MethodName),
						InputType:  proto.String(".testpkg.GetReplyRequest"),
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
