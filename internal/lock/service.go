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

package lock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/form3tech-oss/x-pdb/internal/converters"
	stateclient "github.com/form3tech-oss/x-pdb/internal/state/client"
	statepb "github.com/form3tech-oss/x-pdb/pkg/proto/state/v1"
	"github.com/go-logr/logr"
	"github.com/sourcegraph/conc/pool"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var remoteLockTimeout = 2 * time.Second

// Service is responsible to manage the locks
// used to guarantee that there are no race conditions
// when doing disruptions across clusters.
type Service struct {
	logger          *logr.Logger
	client          client.Client
	reader          client.Reader
	stateClientPool *stateclient.ClientPool
	leaseNamespace  string
	remoteEndpoints []string
}

// NewService creates a new Service instance.
func NewService(
	logger *logr.Logger,
	client client.Client,
	reader client.Reader,
	stateClientPool *stateclient.ClientPool,
	leaseNamespace string,
	remoteEndpoints []string,
) *Service {
	return &Service{
		logger:          logger,
		client:          client,
		reader:          reader,
		stateClientPool: stateClientPool,
		leaseNamespace:  leaseNamespace,
		remoteEndpoints: remoteEndpoints,
	}
}

// Lock creates / updates leases on the local and remote clusters to disallow concurrent pod disruptions for a given namespace and
// selector.
func (s *Service) Lock(ctx context.Context, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) error {
	err := s.LocalLock(ctx, leaseHolderIdentity, namespace, selector)
	if err != nil {
		return fmt.Errorf("unable to lock local cluster: %w", err)
	}

	if len(s.remoteEndpoints) > 0 {
		err = s.remoteLock(ctx, leaseHolderIdentity, namespace, selector)
		if err != nil {
			return fmt.Errorf("unable to lock remote clusters: %w", err)
		}
	}

	return nil
}

// Unlock deletes leases on local and remote clusters to allow pod disruptions to happen.
func (s *Service) Unlock(ctx context.Context, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) error {
	err := s.LocalUnlock(ctx, leaseHolderIdentity, namespace, selector)
	if err != nil {
		return fmt.Errorf("unable to unlock local cluster: %w", err)
	}

	if len(s.remoteEndpoints) > 0 {
		err = s.remoteUnlock(ctx, leaseHolderIdentity, namespace, selector)
		if err != nil {
			return fmt.Errorf("unable to unlock remote clusters: %w", err)
		}
	}

	return nil
}

// LocalLock creates / updates leases on the local cluster to disallow concurrent pod disruptions for a given namespace and selector.
func (s *Service) LocalLock(ctx context.Context, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) error {
	lease := createLeaseForSelector(s.leaseNamespace, leaseHolderIdentity, namespace, selector)
	s.logger.V(2).Info("attempting to lock", "lease", lease.Name, "identity", leaseHolderIdentity)

	err := s.client.Create(ctx, lease)
	// lease already exists, verify lease time and take over if possible.
	if apierrors.IsAlreadyExists(err) {
		err = s.reader.Get(ctx, client.ObjectKeyFromObject(lease), lease)
		if err != nil {
			// ignore the NotFound case and let the caller retry.
			return fmt.Errorf("unable to take over lease: %w", err)
		}

		deadline := lease.Spec.AcquireTime.Add(time.Second * time.Duration(*lease.Spec.LeaseDurationSeconds))
		if deadline.After(time.Now()) {
			s.logger.Info(
				"lease deadline not reached",
				"identity", *lease.Spec.HolderIdentity,
				"acquired", lease.Spec.AcquireTime.String(),
				"durationSeconds", *lease.Spec.LeaseDurationSeconds,
				"expiresSeconds", time.Until(deadline).Seconds())
			return errors.New("lease deadline not reached")
		}

		// the lease timed out, update identity + acquire time
		s.logger.Info("lease deadline reached. Acquiring it.",
			"lease", lease.Name,
			"oldIdentity", lease.Spec.HolderIdentity,
			"newIdentity", leaseHolderIdentity,
			"acquireTime", lease.Spec.AcquireTime.Time.String(),
			"duration", *lease.Spec.LeaseDurationSeconds,
			"now", time.Now().String())

		lease.Spec.AcquireTime = &metav1.MicroTime{Time: time.Now()}
		lease.Spec.HolderIdentity = &leaseHolderIdentity

		err = s.client.Update(ctx, lease)
		if err != nil {
			return fmt.Errorf("unable to update lease: %w", err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to acquire lease: %w", err)
	}

	s.logger.V(2).Info("acquired lock", "lease", lease.Name, "identity", lease.Spec.HolderIdentity)
	return nil
}

// LocalUnlock deletes leases on the local cluster to allow pod disruptions to happen.
func (s *Service) LocalUnlock(ctx context.Context, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) error {
	lease := createLeaseForSelector(s.leaseNamespace, leaseHolderIdentity, namespace, selector)
	err := s.reader.Get(ctx, client.ObjectKeyFromObject(lease), lease)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if *lease.Spec.HolderIdentity != leaseHolderIdentity {
		return fmt.Errorf("holder identity does not match. Someone else has taken over the lease. expected=%q seen=%q ", leaseHolderIdentity, *lease.Spec.HolderIdentity)
	}

	return s.client.Delete(ctx, lease)
}

func (s *Service) remoteLock(ctx context.Context, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) error {
	if len(s.remoteEndpoints) == 0 {
		return nil
	}

	req := &statepb.LockRequest{
		LeaseHolderIdentity: leaseHolderIdentity,
		Namespace:           namespace,
		LabelSelector:       converters.ConvertLabelSelectorToState(selector),
	}

	p := pool.NewWithResults[*statepb.LockResponse]().
		WithErrors().
		WithMaxGoroutines(len(s.remoteEndpoints)).
		WithContext(ctx)

	for _, e := range s.remoteEndpoints {
		p.Go(func(ctx context.Context) (*statepb.LockResponse, error) {
			cli, err := s.stateClientPool.Get(e)
			if err != nil {
				return nil, err
			}

			cctx, cancel := context.WithTimeout(ctx, remoteLockTimeout)
			defer cancel()

			return cli.Lock(cctx, req)
		})
	}

	results, err := p.Wait()
	if err != nil {
		return err
	}

	var errs []error
	for _, r := range results {
		if !r.Acquired {
			errs = append(errs, fmt.Errorf("lock not acquired: %s", r.Error))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (s *Service) remoteUnlock(ctx context.Context, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) error {
	if len(s.remoteEndpoints) == 0 {
		return nil
	}

	req := &statepb.UnlockRequest{
		LeaseHolderIdentity: leaseHolderIdentity,
		Namespace:           namespace,
		LabelSelector:       converters.ConvertLabelSelectorToState(selector),
	}

	p := pool.NewWithResults[*statepb.UnlockResponse]().
		WithErrors().
		WithMaxGoroutines(len(s.remoteEndpoints)).
		WithContext(ctx)

	for _, e := range s.remoteEndpoints {
		p.Go(func(ctx context.Context) (*statepb.UnlockResponse, error) {
			cli, err := s.stateClientPool.Get(e)
			if err != nil {
				return nil, err
			}

			cctx, cancel := context.WithTimeout(ctx, remoteLockTimeout)
			defer cancel()

			return cli.Unlock(cctx, req)
		})
	}

	results, err := p.Wait()
	if err != nil {
		return err
	}

	var errs []error
	for _, r := range results {
		if !r.Unlocked {
			errs = append(errs, fmt.Errorf("lock was not unlocked: %s", r.Error))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
