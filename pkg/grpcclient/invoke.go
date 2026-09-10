package grpcclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Invoke calls the given unary method with an empty request message (apix
// does not yet support sending request bodies), attaching md as outgoing
// gRPC metadata, and returns the response marshalled as pretty-printed JSON.
func Invoke(ctx context.Context, conn *grpc.ClientConn, method *MethodInfo, md map[string]string) ([]byte, error) {
	if method.IsClientStreaming || method.IsServerStreaming {
		return nil, fmt.Errorf("method %q is a streaming RPC; apix grpc currently only supports unary methods", method.FullName)
	}

	if len(md) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))
	}

	req := dynamicpb.NewMessage(method.Input)
	resp := dynamicpb.NewMessage(method.Output)

	if err := conn.Invoke(ctx, method.FullName, req, resp); err != nil {
		return nil, fmt.Errorf("rpc call to %q failed: %w", method.FullName, err)
	}

	out, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return out, nil
}
