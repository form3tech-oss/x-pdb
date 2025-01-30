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

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/form3tech-oss/x-pdb/internal/disruptionprobe"
	"github.com/form3tech-oss/x-pdb/internal/lock"
	"github.com/form3tech-oss/x-pdb/internal/pdb"
	"github.com/form3tech-oss/x-pdb/internal/preactivities"
	stateclient "github.com/form3tech-oss/x-pdb/internal/state/client"
	"github.com/form3tech-oss/x-pdb/internal/webhooks"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	xpdbv1alpha1 "github.com/form3tech-oss/x-pdb/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(xpdbv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var webhookCertsDir string
	var webhookPort int
	var stateCertsDir string
	var remoteEndpoints string
	var leaseNamespace string
	var podID string
	var kubeContext string
	var clusterID string
	var dryRun bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&webhookCertsDir, "webhook-certs-dir", "", "The directory that contains webhook certificates")
	flag.IntVar(&webhookPort, "webhook-port", 9443, "The webhook binding port")
	flag.StringVar(&stateCertsDir, "state-certs-dir", "", "The directory that contains state server certificates")
	flag.StringVar(&remoteEndpoints, "remote-endpoints", "", "The list of endpoints of the remote pdb controllers")
	flag.StringVar(&leaseNamespace, "namespace", "kube-system", "the namespace in which the controller runs in")
	flag.StringVar(&podID, "pod-id", os.Getenv("HOSTNAME"),
		"The ID of the pod x-pdb pod. Used as prefix for the lease-holder-identity to obtain locks across clusters.",
	)
	flag.StringVar(&clusterID, "cluster-id", "no-id-set",
		"The ID of the cluster where x-pdb is running."+
			"Used as prefix for the lease-holder-identity to obtain locks across clusters.",
	)
	flag.StringVar(&kubeContext, "kube-context", "", "kube context to connect to a cluster")
	flag.BoolVar(&dryRun, "dry-run", false,
		"run the admission controller in dry-run mode, which never rejects a voluntary disruption",
	)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	cfg, err := config.GetConfigWithContext(kubeContext)
	if err != nil {
		setupLog.Error(err, "unable to get kubernetes config")
		os.Exit(1)
	}
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	signalHandler := ctrl.SetupSignalHandler()

	cli, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create kubernetes client")
		os.Exit(1)
	}

	remoteEndpointsList, err := parseEndpoints(remoteEndpoints)
	if err != nil {
		setupLog.Error(err, "unable to parse remote endpoints")
		os.Exit(1)
	}

	stateClientPool := stateclient.NewClientPool(signalHandler, &logger, stateCertsDir)

	lockService := lock.NewService(
		&logger,
		mgr.GetClient(),
		mgr.GetAPIReader(),
		stateClientPool,
		leaseNamespace,
		remoteEndpointsList,
	)

	disruptionProbeClientPool := disruptionprobe.NewClientPool(signalHandler, &logger, stateCertsDir)
	disruptionProbeService := disruptionprobe.NewService(&logger, disruptionProbeClientPool)

	scaleFinder := pdb.NewScaleFinder(mgr.GetClient(), cli.DiscoveryClient)
	pdbService := pdb.NewService(logger,
		mgr.GetClient(),
		mgr.GetAPIReader(),
		scaleFinder,
		stateClientPool,
		leaseNamespace,
		remoteEndpointsList)

	preactivitiesService := preactivities.NewService(logger, mgr.GetClient())

	{
		hookServer := &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    webhookPort,
				CertDir: webhookCertsDir,
			},
		}
		if err := mgr.Add(hookServer); err != nil {
			setupLog.Error(err, "unable to create pod mutator webhook server")
			os.Exit(1)
		}
		decoder := admission.NewDecoder(mgr.GetScheme())
		podValidationWebhook := webhooks.NewPodValidationWebhook(
			mgr.GetClient(),
			logger,
			decoder,
			mgr.GetEventRecorderFor("x-pdb"),
			clusterID,
			podID,
			dryRun,
			pdbService,
			lockService,
			disruptionProbeService,
			preactivitiesService,
		)
		hookServer.Register("/validate", &webhook.Admission{Handler: podValidationWebhook})
	}

	// +kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(signalHandler); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func parseEndpoints(endpointString string) ([]string, error) {
	//nolint:prealloc
	var endpoints []string
	var errs []error
	splitEndpoints := strings.Split(endpointString, ",")

	if len(splitEndpoints) == 1 && splitEndpoints[0] == "" {
		return endpoints, nil
	}

	for _, ep := range splitEndpoints {
		sanitizedEndpoint := strings.TrimSpace(ep)

		if sanitizedEndpoint == "" {
			errs = append(errs, fmt.Errorf("endpoint cannot be empty"))
			continue
		}

		endpoints = append(endpoints, sanitizedEndpoint)
	}
	return endpoints, errors.Join(errs...)
}
