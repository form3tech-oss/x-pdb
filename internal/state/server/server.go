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

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"

	"github.com/form3tech-oss/x-pdb/internal/converters"
	"github.com/form3tech-oss/x-pdb/internal/lock"
	"github.com/form3tech-oss/x-pdb/internal/pdb"
	statepb "github.com/form3tech-oss/x-pdb/pkg/protos/state"
	"github.com/go-logr/logr"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Server struct {
	stateServer *stateServer
	logger      *logr.Logger
	port        int
	certsDir    string
}

func NewServer(pdbService *pdb.Service, lockService *lock.Service, logger *logr.Logger, port int, certsDir string) *Server {
	s := &stateServer{
		pdbService:  pdbService,
		lockService: lockService,
		logger:      logger,
	}

	return &Server{
		stateServer: s,
		logger:      logger,
		port:        port,
		certsDir:    certsDir,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting grpc state server", "port", s.port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	cw, err := certwatcher.New(path.Join(s.certsDir, "tls.crt"), path.Join(s.certsDir, "tls.key"))
	if err != nil {
		return fmt.Errorf("error creating cert watcher: %w", err)
	}

	go func() {
		if err := cw.Start(ctx); err != nil {
			s.logger.Error(err, "certificate watcher error")
		}
	}()

	certPool := x509.NewCertPool()
	//nolint:gosec
	clientCABytes, err := os.ReadFile(filepath.Join(s.certsDir, "ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read client CA cert: %w", err)
	}

	ok := certPool.AppendCertsFromPEM(clientCABytes)
	if !ok {
		return fmt.Errorf("failed to append client CA cert to CA pool")
	}

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.01, 0.1, 0.3, 0.6, 1, 3, 5}),
		),
	)
	metrics.Registry.MustRegister(srvMetrics)

	grpcPanicRecoveryHandler := func(p any) (err error) {
		s.logger.Error(fmt.Errorf("recovered from panic"), "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.Creds(
			credentials.NewTLS(
				&tls.Config{
					MinVersion:     tls.VersionTLS13,
					ClientAuth:     tls.RequireAndVerifyClientCert,
					ClientCAs:      certPool,
					GetCertificate: cw.GetCertificate,
				},
			),
		),
	)
	statepb.RegisterStateServer(grpcServer, s.stateServer)

	errCh := make(chan error)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return nil
	case <-errCh:
		return err
	}
}

type stateServer struct {
	pdbService  *pdb.Service
	lockService *lock.Service
	logger      *logr.Logger
	statepb.UnimplementedStateServer
}

func (s *stateServer) Lock(ctx context.Context, req *statepb.LockRequest) (*statepb.LockResponse, error) {
	labelSelector := converters.ConvertLabelSelectorToMetaV1(req.LabelSelector)

	resp := &statepb.LockResponse{}
	err := s.lockService.LocalLock(ctx, req.LeaseHolderIdentity, req.Namespace, labelSelector)
	if err == nil {
		resp.Acquired = true
	} else {
		s.logger.Error(err, "unable to lock xpdb")
		resp.Error = err.Error()
	}

	return resp, nil
}

func (s *stateServer) Unlock(ctx context.Context, req *statepb.UnlockRequest) (*statepb.UnlockResponse, error) {
	labelSelector := converters.ConvertLabelSelectorToMetaV1(req.LabelSelector)

	resp := &statepb.UnlockResponse{}
	err := s.lockService.LocalUnlock(context.Background(), req.LeaseHolderIdentity, req.Namespace, labelSelector)
	if err == nil {
		resp.Unlocked = true
	} else {
		s.logger.Error(err, "unable to unlock xpdb")
		resp.Error = err.Error()
	}

	return resp, nil
}

func (s *stateServer) GetState(ctx context.Context, req *statepb.GetStateRequest) (*statepb.GetStateResponse, error) {
	labelSelector := converters.ConvertLabelSelectorToMetaV1(req.LabelSelector)

	desiredHealthy, healthy, err := s.pdbService.GetPodCounts(context.Background(), req.Namespace, labelSelector)
	if err != nil {
		s.logger.Error(err, "unable to get pod counts")
		return nil, status.Errorf(codes.Internal, "unable to get pod counts")
	}

	return &statepb.GetStateResponse{
		DesiredHealthy: desiredHealthy,
		Healthy:        healthy,
	}, nil
}
