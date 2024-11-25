package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	disruptionprobepb "github.com/form3tech-oss/x-pdb/pkg/protos/disruptionprobe"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()
	logger = ctrl.Log
)

func init() {
	log.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var listenAddr string
	var certsDir string
	var namespace string

	flag.StringVar(&listenAddr, "listen-addr", ":8080", "the listen address of the grpc server")
	flag.StringVar(&certsDir, "certs-dir", "/etc/certs", "The directory that contains the grpc certificates")
	flag.StringVar(&namespace, "namespace", "", "The namespace where the app is installed")
	flag.Parse()

	cfg := config.GetConfigOrDie()
	ctrlClient, err := client.NewWithWatch(cfg, client.Options{})
	if err != nil {
		logger.Error(err, "unable to create kubernetes client")
		os.Exit(1)
	}

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	configWatcher := &configWatcher{
		ctx:                   ctx,
		namespace:             namespace,
		notAllowedDisruptions: map[string]bool{},
		mux:                   &sync.Mutex{},
		ctrlClient:            ctrlClient,
	}

	srv := server{
		ctx:        ctx,
		listenAddr: listenAddr,
		certsDir:   certsDir,
		srv: &disruptionProbeServer{
			configWatcher: configWatcher,
		},
	}

	g.Go(configWatcher.Start)
	g.Go(srv.Start)
	g.Go(handleTerminationSignal(ctx))

	err = g.Wait()
	if err != nil {
		logger.Error(err, "waiter finished with error")
		os.Exit(1)
	}
	logger.Info("waiter finished")
}

func handleTerminationSignal(ctx context.Context) func() error {
	return func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-ctx.Done():
			err := ctx.Err()
			if err == context.Canceled {
				return nil
			}
			return err
		}
	}
}

type configWatcher struct {
	ctx                   context.Context
	ctrlClient            client.WithWatch
	namespace             string
	notAllowedDisruptions map[string]bool
	mux                   *sync.Mutex
}

func (w *configWatcher) Start() error {
	var configMapList corev1.ConfigMapList
	watcher, err := w.ctrlClient.Watch(w.ctx, &configMapList, client.InNamespace(w.namespace), client.MatchingLabels{
		"test-disruption-probe-config": "true",
	})
	if err != nil {
		logger.Error(err, "unable to create watcher for ConfigMaps")
		return err
	}

	cm := &corev1.ConfigMap{}
	err = w.ctrlClient.Get(w.ctx, types.NamespacedName{Namespace: w.namespace, Name: "test-disruption-probe-config"}, cm)
	if err != nil {
		logger.Error(err, "unable to get config configmap")
		return err
	}

	err = w.parseConfigMap(cm)
	if err != nil {
		logger.Error(err, "unable to parse configmap")
		return err
	}

	for {
		select {
		case obj := <-watcher.ResultChan():
			logger.Info("received object", "obj", obj)
			cm, ok := obj.Object.(*corev1.ConfigMap)
			if !ok {
				logger.Error(fmt.Errorf("unable to convert object to configmap"), "unable to convert object to configmap")
				continue
			}
			err := w.parseConfigMap(cm)
			if err != nil {
				logger.Error(err, "could not parse data on configmap")
			}
		case <-w.ctx.Done():
			return nil
		}
	}
}

func (w *configWatcher) parseConfigMap(cm *corev1.ConfigMap) error {
	content, ok := cm.Data["notAllowedDisruptions"]
	if !ok {
		logger.Error(fmt.Errorf("configmap didn't have the config property"), "configmap didn't have the config property")
	}

	items := strings.Split(string(content), ",")

	disruptionAllowed := map[string]bool{}
	for i := range items {
		disruptionAllowed[items[i]] = true
	}

	w.mux.Lock()
	defer w.mux.Unlock()
	w.notAllowedDisruptions = disruptionAllowed

	return nil
}

func (r *configWatcher) IsDisruptionAllowed(namespace, name string) bool {
	r.mux.Lock()
	defer r.mux.Unlock()

	return !r.notAllowedDisruptions[fmt.Sprintf("%s/%s", namespace, name)]
}

type server struct {
	ctx        context.Context
	listenAddr string
	certsDir   string
	srv        *disruptionProbeServer
}

func (s *server) Start() error {
	lis, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}

	cw, err := certwatcher.New(path.Join(s.certsDir, "tls.crt"), path.Join(s.certsDir, "tls.key"))
	if err != nil {
		return fmt.Errorf("error creating cert watcher: %w", err)
	}

	go func() {
		if err := cw.Start(s.ctx); err != nil {
			logger.Error(err, "certificate watcher error")
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

	grpcServer := grpc.NewServer(
		grpc.Creds(
			credentials.NewTLS(
				&tls.Config{
					MinVersion:     tls.VersionTLS13,
					ClientCAs:      certPool,
					GetCertificate: cw.GetCertificate,
				},
			),
		),
	)
	disruptionprobepb.RegisterDisruptionProbeServer(grpcServer, s.srv)

	logger.Info(fmt.Sprintf("starting out grpc server on %s", s.listenAddr))
	errCh := make(chan error)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case <-s.ctx.Done():
		grpcServer.GracefulStop()
		return nil
	case <-errCh:
		return err
	}
}

type disruptionProbeServer struct {
	configWatcher *configWatcher
	disruptionprobepb.UnimplementedDisruptionProbeServer
}

func (s *disruptionProbeServer) IsDisruptionAllowed(ctx context.Context, req *disruptionprobepb.IsDisruptionAllowedRequest) (*disruptionprobepb.IsDisruptionAllowedResponse, error) {
	logger.Info("received request", "namespace", req.PodNamespace, "name", req.PodName)

	isDisruptionAllowed := s.configWatcher.IsDisruptionAllowed(req.PodNamespace, req.PodName)

	return &disruptionprobepb.IsDisruptionAllowedResponse{
		IsAllowed: isDisruptionAllowed,
		Error:     "",
	}, nil
}
