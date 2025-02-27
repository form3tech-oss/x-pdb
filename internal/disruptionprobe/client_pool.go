package disruptionprobe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/form3tech-oss/x-pdb/internal/metrics"
	disruptionprobepb "github.com/form3tech-oss/x-pdb/pkg/proto/disruptionprobe/v1"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientPool struct {
	clients  map[string]disruptionprobepb.DisruptionProbeServiceClient
	mux      *sync.Mutex
	ctx      context.Context
	logger   *logr.Logger
	certsDir string
}

func NewClientPool(
	ctx context.Context,
	logger *logr.Logger,
	certsDir string,
) *ClientPool {
	return &ClientPool{
		certsDir: certsDir,
		clients:  make(map[string]disruptionprobepb.DisruptionProbeServiceClient),
		ctx:      ctx,
		mux:      &sync.Mutex{},
		logger:   logger,
	}
}

func (p *ClientPool) Get(endpoint string) (disruptionprobepb.DisruptionProbeServiceClient, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	c, found := p.clients[endpoint]
	if !found {
		c, err := p.newClient(endpoint)
		if err != nil {
			return nil, err
		}

		p.clients[endpoint] = c
		return c, nil
	}

	return c, nil
}

func (p *ClientPool) newClient(endpoint string) (disruptionprobepb.DisruptionProbeServiceClient, error) {
	certPool := x509.NewCertPool()

	p.logger.Info("certs dir", "certsdir", p.certsDir)
	//nolint:gosec
	clientCABytes, err := os.ReadFile(filepath.Join(p.certsDir, "ca.crt"))
	if err != nil {
		return nil, fmt.Errorf("failed to read client CA cert: %w", err)
	}

	ok := certPool.AppendCertsFromPEM(clientCABytes)
	if !ok {
		return nil, fmt.Errorf("failed to append client CA cert to CA pool")
	}

	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithChainUnaryInterceptor(
			metrics.GrpcClientMetrics.UnaryClientInterceptor(),
		),
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					RootCAs:    certPool,
					MinVersion: tls.VersionTLS12,
				},
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return disruptionprobepb.NewDisruptionProbeServiceClient(conn), nil
}
