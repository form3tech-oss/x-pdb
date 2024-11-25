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
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	xpdbv1alpha1 "github.com/form3tech-oss/x-pdb/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	log.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(xpdbv1alpha1.AddToScheme(scheme))
}

// This test application is used in integration tests.
// It runs as a pod and exposes a readiness probe http server.
// The test application response to a readiness probe is configurable
// through a Kind=ConfigMap, which contains the pod name as key and the readiness
// value as a boolean.
// This application watches the configmap for changes and
// modifies the readiness probe response.
func main() {
	cfg := config.GetConfigOrDie()
	ctrlClient, err := client.NewWithWatch(cfg, client.Options{})
	if err != nil {
		setupLog.Error(err, "unable to create kubernetes client")
		os.Exit(1)
	}

	// respond to SIGINT
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	var reportPodReady bool
	srv := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reportPodReady {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}),
	}
	go func() {
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			setupLog.Error(err, "ListenAndServe() errored")
		}
	}()

	myNamespace := os.Getenv("NAMESPACE")
	if myNamespace == "" {
		setupLog.Info("unable to get namespace of this pod")
		os.Exit(1)
	}

	var configMapList v1.ConfigMapList
	watcher, err := ctrlClient.Watch(ctx, &configMapList, client.InNamespace(myNamespace), client.MatchingLabels{
		"app": "test",
	})
	if err != nil {
		setupLog.Error(err, "unable to create watcher for ConfigMaps")
		os.Exit(1)
	}

	for {
		select {
		// when the configmap changes recalculate pod readiness
		// and modify the `reportPodReady` variable.
		case obj := <-watcher.ResultChan():
			setupLog.Info("received object", "obj", obj)
			cm, ok := obj.Object.(*v1.ConfigMap)
			if !ok {
				setupLog.Error(err, "unable to")
			}
			reportPodReady, err = isPodReady(cm)
			if err != nil {
				setupLog.Error(err, "unable to check pod readiness")
				continue
			}
		// when we receive a shutdown signal shutdown and cleanup
		case <-sigChan:
			setupLog.Info("shutting down")
			watcher.Stop()
			if err = srv.Close(); err != nil {
				setupLog.Error(err, "error closing listener")
			}
			cancel()
		}
	}
}

func isPodReady(cm *v1.ConfigMap) (bool, error) {
	if cm == nil {
		return false, nil
	}
	// when this is set then all pods are always ready.
	if val := cm.Data["always"]; val == "true" {
		return true, nil
	}
	hostname := os.Getenv("HOSTNAME")
	val, ok := cm.Data[hostname]
	if !ok {
		return false, nil
	}
	return strconv.ParseBool(val)
}
