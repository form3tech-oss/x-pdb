package preactivities

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PreActivityAnnotationNamePrefix    = "pre-activity.xpdb.form3.tech/"
	HasPendingDisruptionAnnotationName = "xpdb.form3.tech/has-pending-disruption"
)

type Service struct {
	logger logr.Logger
	cli    client.Client
}

func NewService(
	logger logr.Logger,
	cli client.Client,
) *Service {
	return &Service{
		logger: logger,
		cli:    cli,
	}
}

func (s *Service) CanPodBeDisrupted(ctx context.Context, pod *corev1.Pod) (bool, error) {
	pendingPreactivities := s.getPendingPreactivities(pod)
	if len(pendingPreactivities) == 0 {
		return true, nil
	}

	s.logger.WithValues(
		"podName", pod.Name,
		"podNamespace", pod.Namespace,
		"pendingActivities", pendingPreactivities,
	).Info("pod has pending activities")

	err := s.setPendingDisruptionAnnotationOnPod(ctx, pod)
	if err != nil {
		s.logger.WithValues(
			"podName", pod.Name,
			"podNamespace", pod.Namespace,
		).Error(err, "could not add pending disruption annotation to pod")

		return false, fmt.Errorf("could not add pending disruption annotation to pod")
	}

	return false, nil
}

func (s *Service) getPendingPreactivities(pod *corev1.Pod) []string {
	preactivities := []string{}
	for k := range pod.Annotations {
		if strings.Contains(k, PreActivityAnnotationNamePrefix) {
			preactivity := strings.TrimPrefix(k, PreActivityAnnotationNamePrefix)
			preactivities = append(preactivities, preactivity)
		}
	}
	return preactivities
}

func (s *Service) setPendingDisruptionAnnotationOnPod(ctx context.Context, pod *corev1.Pod) error {
	pod.Annotations[HasPendingDisruptionAnnotationName] = "true"
	err := s.cli.Update(ctx, pod)
	if err != nil {
		return err
	}
	return nil
}
