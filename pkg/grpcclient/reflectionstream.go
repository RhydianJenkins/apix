package grpcclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// reflectionStream wraps the bidi-streaming ServerReflectionInfo RPC as a
// simple request/response helper, since apix only ever sends one request at
// a time and waits for its matching response.
type reflectionStream struct {
	stream reflectionpb.ServerReflection_ServerReflectionInfoClient
}

func newReflectionStream(ctx context.Context, conn *grpc.ClientConn) (*reflectionStream, error) {
	stream, err := reflectionpb.NewServerReflectionClient(conn).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open reflection stream (does the server have reflection enabled?): %w", err)
	}

	return &reflectionStream{stream: stream}, nil
}

func (r *reflectionStream) close() {
	_ = r.stream.CloseSend()
}

func (r *reflectionStream) listServices() ([]string, error) {
	resp, err := r.send(&reflectionpb.ServerReflectionRequest{
		MessageRequest: &reflectionpb.ServerReflectionRequest_ListServices{},
	})
	if err != nil {
		return nil, err
	}

	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, fmt.Errorf("unexpected reflection response for ListServices: %v", resp)
	}

	services := make([]string, 0, len(listResp.GetService()))
	for _, s := range listResp.GetService() {
		services = append(services, s.GetName())
	}

	return services, nil
}

// symbol is a type, service, or method full name. The response includes its
// transitive dependencies, but not ones already returned earlier on this
// stream.
func (r *reflectionStream) fileContainingSymbol(symbol string) ([]*descriptorpb.FileDescriptorProto, error) {
	resp, err := r.send(&reflectionpb.ServerReflectionRequest{
		MessageRequest: &reflectionpb.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: symbol,
		},
	})
	if err != nil {
		return nil, err
	}

	fdResp := resp.GetFileDescriptorResponse()
	if fdResp == nil {
		return nil, fmt.Errorf("unexpected reflection response for FileContainingSymbol(%q): %v", symbol, resp)
	}

	files := make([]*descriptorpb.FileDescriptorProto, 0, len(fdResp.GetFileDescriptorProto()))
	for _, raw := range fdResp.GetFileDescriptorProto() {
		fd := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(raw, fd); err != nil {
			return nil, fmt.Errorf("failed to unmarshal file descriptor for %q: %w", symbol, err)
		}
		files = append(files, fd)
	}

	return files, nil
}

func (r *reflectionStream) send(req *reflectionpb.ServerReflectionRequest) (*reflectionpb.ServerReflectionResponse, error) {
	if err := r.stream.Send(req); err != nil {
		return nil, fmt.Errorf("failed to send reflection request: %w", err)
	}

	resp, err := r.stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive reflection response: %w", err)
	}

	if errResp := resp.GetErrorResponse(); errResp != nil {
		return nil, fmt.Errorf("reflection error: %s (code %d)", errResp.GetErrorMessage(), errResp.GetErrorCode())
	}

	return resp, nil
}
