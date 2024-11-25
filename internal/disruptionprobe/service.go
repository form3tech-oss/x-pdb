package disruptionprobe

import (
	"context"
	"errors"
	"time"

	xpdbv1alpha1 "github.com/form3tech-oss/x-pdb/api/v1alpha1"
	"github.com/form3tech-oss/x-pdb/pkg/protos/disruptionprobe"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

var (
	probeTimeout = 2 * time.Second
)

type Service struct {
	clientPool *ClientPool
	logger     *logr.Logger
}

func NewService(
	logger *logr.Logger,
	clientPool *ClientPool,
) *Service {
	return &Service{
		logger:     logger,
		clientPool: clientPool,
	}
}

func (s *Service) CanPodBeDisrupted(ctx context.Context, pod *corev1.Pod, xpdb *xpdbv1alpha1.XPodDisruptionBudget) (bool, error) {
	if xpdb.Spec.Probe == nil {
		return true, nil
	}

	c, err := s.clientPool.Get(xpdb.Spec.Probe.Endpoint)
	if err != nil {
		return false, err
	}

	req := &disruptionprobe.IsDisruptionAllowedRequest{
		PodName:       pod.Name,
		PodNamespace:  pod.Namespace,
		XpdbName:      xpdb.Name,
		XpdbNamespace: xpdb.Namespace,
	}

	cctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	resp, err := c.IsDisruptionAllowed(cctx, req)
	if err != nil {
		return false, err
	}

	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}

	return resp.IsAllowed, nil
}
