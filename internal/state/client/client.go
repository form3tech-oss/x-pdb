/*
Copyright 2024 Form3.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/form3tech-oss/x-pdb/internal/metrics"
	statepb "github.com/form3tech-oss/x-pdb/pkg/protos/state"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
)

type ClientPool struct {
	clients  map[string]statepb.StateClient
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
		ctx:      ctx,
		mux:      &sync.Mutex{},
		logger:   logger,
		certsDir: certsDir,
		clients:  map[string]statepb.StateClient{},
	}
}

func (p *ClientPool) Get(endpoint string) (statepb.StateClient, error) {
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

func (p *ClientPool) newClient(endpoint string) (statepb.StateClient, error) {
	cw, err := certwatcher.New(path.Join(p.certsDir, "tls.crt"), path.Join(p.certsDir, "tls.key"))
	if err != nil {
		return nil, fmt.Errorf("error creating cert watcher: %w", err)
	}

	go func() {
		if err := cw.Start(p.ctx); err != nil {
			p.logger.Error(err, "certificate watcher error")
		}
	}()

	certPool := x509.NewCertPool()
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
					GetClientCertificate: func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
						return cw.GetCertificate(nil)
					},
				},
			),
		),
	)

	if err != nil {
		return nil, err
	}

	return statepb.NewStateClient(conn), nil
}
