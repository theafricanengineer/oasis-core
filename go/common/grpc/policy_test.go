package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/oasislabs/ekiden/go/common/accessctl"
	memorySigner "github.com/oasislabs/ekiden/go/common/crypto/signature/signers/memory"
	"github.com/oasislabs/ekiden/go/common/identity"
)

var (
	_ PingServer = (*pingServer)(nil)
	_ PingClient = (*pingClient)(nil)
)

func CreateCertificate(t *testing.T) (*tls.Certificate, *x509.Certificate) {
	require := require.New(t)

	dataDir, err := ioutil.TempDir("", "ekiden-common-grpc-test_")
	require.NoError(err, "Failed to create a temporary directory")
	defer os.RemoveAll(dataDir)

	ident, err := identity.LoadOrGenerate(dataDir, memorySigner.NewFactory())
	require.NoError(err, "Failed to generate a new identity")
	require.Len(ident.TLSCertificate.Certificate, 1, "The generated identity contains more than 1 TLS certificate in the chain")

	x509Cert, err := x509.ParseCertificate(ident.TLSCertificate.Certificate[0])
	require.NoError(err, "Failed to parse X.509 certificate from TLS certificate")

	return ident.TLSCertificate, x509Cert
}

func connectToGrpcServer(
	ctx context.Context,
	t *testing.T,
	address string,
	creds credentials.TransportCredentials,
) *grpc.ClientConn {
	require := require.New(t)
	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(&CBORCodec{})),
	)
	require.NoErrorf(err, "Failed to connect to the gRPC server: %v", err)
	return conn
}

type PingQuery struct {
}

type PingResponse struct {
}

type PingServer interface {
	Ping(context.Context, *PingQuery) (*PingResponse, error)
}

type pingServer struct {
	policy accessctl.Policy
}

func (s *pingServer) Ping(ctx context.Context, query *PingQuery) (*PingResponse, error) {
	if err := CheckAllowed(ctx, s.policy, "Ping"); err != nil {
		return nil, errors.Wrap(err, "ping: access policy forbade access")
	}
	return &PingResponse{}, nil
}

type PingClient interface {
	Ping(ctx context.Context, in *PingQuery, opts ...grpc.CallOption) (*PingResponse, error)
}

type pingClient struct {
	cc *grpc.ClientConn
}

func (c *pingClient) Ping(ctx context.Context, in *PingQuery, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/PingService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var serviceDesc = grpc.ServiceDesc{
	ServiceName: "PingService",
	HandlerType: (*PingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    pingHandler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

func pingHandler( // nolint: golint
	srv interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	pq := new(PingQuery)
	if err := dec(pq); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingServer).Ping(ctx, pq)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/PingServer/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingServer).Ping(ctx, req.(*PingQuery))
	}
	return interceptor(ctx, pq, info, handler)
}

func TestAccessPolicy(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	host := "localhost"
	var port uint16 = 50123

	serverTLSCert, serverX509Cert := CreateCertificate(t)
	clientTLSCert, clientX509Cert := CreateCertificate(t)

	serverCertPool := x509.NewCertPool()
	serverCertPool.AddCert(serverX509Cert)

	// Create a new gRPC server.
	grpcServer, err := NewServerTCP(host, port, serverTLSCert, []grpc.ServerOption{grpc.CustomCodec(&CBORCodec{})})
	require.NoErrorf(err, "Failed to create a new gRPC server: %v", err)
	// Create an empty access control policy and register it with the ping service.
	policy := accessctl.NewPolicy()
	grpcServer.Server().RegisterService(&serviceDesc, &pingServer{policy: policy})
	// Start gRPC server in a separate goroutine.
	err = grpcServer.Start()
	require.NoErrorf(err, "Failed to start the gRPC server: %v", err)

	clientTLSCredsWithoutCert := credentials.NewTLS(&tls.Config{
		RootCAs:    serverCertPool,
		ServerName: "ekiden-node",
	})
	clientTLSCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{*clientTLSCert},
		RootCAs:      serverCertPool,
		ServerName:   "ekiden-node",
	})
	address := fmt.Sprintf("%s:%d", host, port)

	// Connect to the gRPC server without a client certificate.
	conn := connectToGrpcServer(ctx, t, address, clientTLSCredsWithoutCert)
	defer conn.Close()
	// Create a new ping client.
	client := &pingClient{conn}

	_, err = client.Ping(ctx, &PingQuery{})
	require.EqualError(
		err,
		"rpc error: code = Unknown desc = ping: access policy forbade access: grpc: unexpected number of peer certificates: 0",
		"Calling Ping without a client certificate should not be allowed",
	)

	// Connect to the gRPC server with a client certificate.
	conn = connectToGrpcServer(ctx, t, address, clientTLSCreds)
	defer conn.Close()
	// Create a new ping client.
	client = &pingClient{conn}

	_, err = client.Ping(ctx, &PingQuery{})
	require.Error(err, "Calling Ping with an empty access policy should not be allowed")
	require.Contains(
		err.Error(),
		"rpc error: code = Unknown desc = ping: access policy forbade access: grpc: calling Ping method not allowed for client",
		"Calling Ping with an empty access policy should not be allowed",
	)

	// Add a policy rule to allow the client to call Ping.
	subject := accessctl.SubjectFromCertificate(clientX509Cert)
	policy.Allow(subject, "Ping")

	res, err := client.Ping(ctx, &PingQuery{})
	require.NoError(err, "Calling Ping with proper access policy set should succeed")
	require.IsType(&PingResponse{}, res, "Calling Ping should return a response of the correct type")
}