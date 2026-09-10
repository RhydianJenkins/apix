package grpcclient

import (
	"bytes"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Invoke calls method on conn. reqBody, if non-nil and non-blank, is
// treated as a JSON-encoded request message (protojson-compatible field
// names) and unmarshalled into the RPC's input message; a nil or
// whitespace-only reqBody results in an empty request message, same as
// before request bodies were supported.
func Invoke(ctx context.Context, conn *grpc.ClientConn, method *MethodInfo, reqBody *[]byte, md map[string]string) ([]byte, error) {
	if method.IsClientStreaming || method.IsServerStreaming {
		return nil, fmt.Errorf("method %q is a streaming RPC; apix grpc currently only supports unary methods", method.FullName)
	}

	if len(md) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))
	}

	req := dynamicpb.NewMessage(method.Input)

	if reqBody != nil && len(bytes.TrimSpace(*reqBody)) > 0 {
		if err := protojson.Unmarshal(*reqBody, req); err != nil {
			return nil, fmt.Errorf("invalid JSON request body for %q: %w", method.FullName, err)
		}
	}

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
