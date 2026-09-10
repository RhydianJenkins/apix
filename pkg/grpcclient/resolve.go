package grpcclient

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// reflectionServiceNames are excluded from method resolution - they're the
// reflection service itself, not something a user would want to call.
var reflectionServiceNames = map[string]bool{
	"grpc.reflection.v1.ServerReflection":      true,
	"grpc.reflection.v1alpha.ServerReflection": true,
}

type MethodInfo struct {
	FullName          string
	Input             protoreflect.MessageDescriptor
	Output            protoreflect.MessageDescriptor
	IsClientStreaming bool
	IsServerStreaming bool
}

// methodArg may be a bare method name (e.g. "GetUser") or a qualified
// "package.Service/Method" path. md is sent as outgoing metadata on the
// reflection stream itself, since servers that require auth for regular
// RPCs typically require it for reflection too.
func ResolveMethod(ctx context.Context, conn *grpc.ClientConn, methodArg string, md map[string]string) (*MethodInfo, error) {
	files, serviceNames, err := discoverServices(ctx, conn, md)
	if err != nil {
		return nil, err
	}

	return findMethod(files, serviceNames, methodArg)
}

// ListMethods returns every non-reflection method available on conn, as
// fully qualified "package.Service/Method" names sorted alphabetically.
// It's intended for shell completion, where each candidate must be a
// directly usable, unambiguous `apix grpc` argument.
func ListMethods(ctx context.Context, conn *grpc.ClientConn, md map[string]string) ([]string, error) {
	files, serviceNames, err := discoverServices(ctx, conn, md)
	if err != nil {
		return nil, err
	}

	var names []string

	for _, svcName := range serviceNames {
		if reflectionServiceNames[svcName] {
			continue
		}

		desc, err := files.FindDescriptorByName(protoreflect.FullName(svcName))
		if err != nil {
			continue
		}

		svcDesc, ok := desc.(protoreflect.ServiceDescriptor)
		if !ok {
			continue
		}

		methods := svcDesc.Methods()
		for i := 0; i < methods.Len(); i++ {
			names = append(names, fmt.Sprintf("%s/%s", svcName, methods.Get(i).Name()))
		}
	}

	sort.Strings(names)

	return names, nil
}

// discoverServices opens a reflection stream, lists the server's services,
// and fetches+builds the file descriptors needed to inspect their methods.
// md is sent as outgoing metadata on the reflection stream itself, since
// servers that require auth for regular RPCs typically require it for
// reflection too.
func discoverServices(ctx context.Context, conn *grpc.ClientConn, md map[string]string) (*protoregistry.Files, []string, error) {
	if len(md) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(md))
	}

	rs, err := newReflectionStream(ctx, conn)
	if err != nil {
		return nil, nil, err
	}
	defer rs.close()

	serviceNames, err := rs.listServices()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list services via reflection: %w", err)
	}

	fileProtos := make(map[string]*descriptorpb.FileDescriptorProto)

	for _, svc := range serviceNames {
		if reflectionServiceNames[svc] {
			continue
		}

		files, err := rs.fileContainingSymbol(svc)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch descriptors for service %q: %w", svc, err)
		}

		for _, fd := range files {
			fileProtos[fd.GetName()] = fd
		}
	}

	fdSet := &descriptorpb.FileDescriptorSet{}
	for _, fd := range fileProtos {
		fdSet.File = append(fdSet.File, fd)
	}

	files, err := protodesc.NewFiles(fdSet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build descriptors from reflection response: %w", err)
	}

	return files, serviceNames, nil
}

func findMethod(files *protoregistry.Files, serviceNames []string, methodArg string) (*MethodInfo, error) {
	qualifiedService, qualifiedMethod, isQualified := splitServiceMethod(methodArg)

	var candidates []*MethodInfo
	var candidateNames []string

	for _, svcName := range serviceNames {
		if reflectionServiceNames[svcName] {
			continue
		}

		if isQualified && svcName != qualifiedService {
			continue
		}

		desc, err := files.FindDescriptorByName(protoreflect.FullName(svcName))
		if err != nil {
			continue
		}

		svcDesc, ok := desc.(protoreflect.ServiceDescriptor)
		if !ok {
			continue
		}

		methods := svcDesc.Methods()
		for i := 0; i < methods.Len(); i++ {
			m := methods.Get(i)

			wantName := methodArg
			if isQualified {
				wantName = qualifiedMethod
			}

			if string(m.Name()) != wantName {
				continue
			}

			candidates = append(candidates, &MethodInfo{
				FullName:          fmt.Sprintf("/%s/%s", svcName, m.Name()),
				Input:             m.Input(),
				Output:            m.Output(),
				IsClientStreaming: m.IsStreamingClient(),
				IsServerStreaming: m.IsStreamingServer(),
			})
			candidateNames = append(candidateNames, fmt.Sprintf("%s/%s", svcName, m.Name()))
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no method named %q found via server reflection", methodArg)
	}

	if len(candidates) > 1 {
		sort.Strings(candidateNames)
		return nil, fmt.Errorf(
			"method name %q is ambiguous, found on multiple services: %s\nqualify it, e.g. `apix grpc %s`",
			methodArg, strings.Join(candidateNames, ", "), candidateNames[0],
		)
	}

	return candidates[0], nil
}

// ok is false if arg has no "/" (i.e. it's a bare method name).
func splitServiceMethod(arg string) (service, method string, ok bool) {
	idx := strings.LastIndex(arg, "/")
	if idx < 0 {
		return "", "", false
	}

	return arg[:idx], arg[idx+1:], true
}
