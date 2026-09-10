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
}

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
	s.grpcSrv = grpc.NewServer(grpc.StreamInterceptor(s.streamInterceptor))

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

func (s *Server) Serve() error {
	return s.grpcSrv.Serve(s.Listener)
}

func (s *Server) Addr() string {
	return s.Listener.Addr().String()
}

// Host returns the host part of Addr, with no port - suitable for a
// config.Domain.Base, which must never include a port itself.
func (s *Server) Host() string {
	host, _, err := net.SplitHostPort(s.Addr())
	if err != nil {
		return s.Addr()
	}

	return host
}

// Port returns the numeric port part of Addr, for use as
// config.GRPCOptions.Port.
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

// streamInterceptor records incoming metadata for streaming RPCs (e.g. the
// reflection service's ServerReflectionInfo), so tests can assert that
// metadata reaches reflection calls, not just unary ones.
func (s *Server) streamInterceptor(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
		s.lastMetadata = md
	}

	return handler(srv, ss)
}

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
