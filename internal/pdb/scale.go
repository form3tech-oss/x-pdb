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

package pdb

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	controllerKindRS  = extensionsv1beta1.SchemeGroupVersion.WithKind("ReplicaSet")
	controllerKindSS  = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
	controllerKindRC  = corev1.SchemeGroupVersion.WithKind("ReplicationController")
	controllerKindDep = extensionsv1beta1.SchemeGroupVersion.WithKind("Deployment")
)

// controllerAndScale is used to return (controller, scale) pairs from the
// controller finder functions.
type controllerAndScale struct {
	types.UID
	scale int32
}

// podControllerFinder is a function type that maps a pod to a list of
// controllers and their scale.
type podControllerFinder func(ctx context.Context, controllerRef *metav1.OwnerReference, namespace string) (*controllerAndScale, error)

// ScaleFinder implements the business logic
// to find the `scale` sub resource of a pod.
type ScaleFinder struct {
	client          client.Client
	discoveryClient *discovery.DiscoveryClient
}

// NewScaleFinder instantiates a new ScaleFinder.
func NewScaleFinder(client client.Client, discoveryClient *discovery.DiscoveryClient) *ScaleFinder {
	return &ScaleFinder{
		client:          client,
		discoveryClient: discoveryClient,
	}
}

// FindExpectedScale returns the expected scale for the given pods.
func (s *ScaleFinder) FindExpectedScale(ctx context.Context, pods []*corev1.Pod) (expectedCount int32, unmanagedPods []string, err error) {
	// When the user specifies a fraction of pods that must be available, we
	// use as the fraction's denominator
	// SUM_{all c in C} scale(c)
	// where C is the union of C_p1, C_p2, ..., C_pN
	// and each C_pi is the set of controllers controlling the pod pi

	// k8s only defines what will happens when 0 or 1 controllers control a
	// given pod.  We explicitly exclude the 0 controllers case here, and we
	// report an error if we find a pod with more than 1 controller.  Thus in
	// practice each C_pi is a set of exactly 1 controller.

	// A mapping from controllers to their scale.
	controllerScale := map[types.UID]int32{}

	// 1. Find the controller for each pod.

	// As of now, we allow PDBs to be applied to pods via selectors, so there
	// can be unmanaged pods(pods that don't have backing controllers) but still have PDBs associated.
	// Such pods are to be collected and PDB backing them should be enqueued instead of immediately throwing
	// a sync error. This ensures disruption controller is not frequently updating the status subresource and thus
	// preventing excessive and expensive writes to etcd.
	// With ControllerRef, a pod can only have 1 controller.
	for _, pod := range pods {
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef == nil {
			unmanagedPods = append(unmanagedPods, pod.Name)
			continue
		}

		// If we already know the scale of the controller there is no need to do anything.
		if _, found := controllerScale[controllerRef.UID]; found {
			continue
		}

		// Check all the supported controllers to find the desired scale.
		foundController := false
		for _, finder := range s.finders() {
			var controllerNScale *controllerAndScale
			controllerNScale, err = finder(ctx, controllerRef, pod.Namespace)
			if err != nil {
				return
			}
			if controllerNScale != nil {
				controllerScale[controllerNScale.UID] = controllerNScale.scale
				foundController = true
				break
			}
		}
		if !foundController {
			err = fmt.Errorf("found no controllers for pod %q: %s", pod.Name, pod.String())
			return
		}
	}

	// 2. Add up all the controllers.
	expectedCount = 0
	for _, count := range controllerScale {
		expectedCount += count
	}

	return expectedCount, unmanagedPods, err
}

func (s *ScaleFinder) finders() []podControllerFinder {
	return []podControllerFinder{s.getPodReplicationController, s.getPodDeployment, s.getPodReplicaSet, s.getPodStatefulSet, s.getScaleController}
}

func (s *ScaleFinder) getPodReplicaSet(ctx context.Context, controllerRef *metav1.OwnerReference, namespace string) (*controllerAndScale, error) {
	ok, err := verifyGroupKind(controllerRef, controllerKindRS.Kind, []string{"apps", "extensions"})
	if !ok || err != nil {
		return nil, err
	}
	var rs appsv1.ReplicaSet
	err = s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: controllerRef.Name}, &rs)
	if err != nil {
		// The only possible error is NotFound, which is ok here.
		return nil, nil
	}
	if rs.UID != controllerRef.UID {
		return nil, nil
	}
	controllerRef = metav1.GetControllerOf(&rs)
	if controllerRef != nil && controllerRef.Kind == controllerKindDep.Kind {
		// Skip RS if it's controlled by a Deployment.
		return nil, nil
	}
	return &controllerAndScale{rs.UID, *(rs.Spec.Replicas)}, nil
}

func (s *ScaleFinder) getPodStatefulSet(ctx context.Context, controllerRef *metav1.OwnerReference, namespace string) (*controllerAndScale, error) {
	ok, err := verifyGroupKind(controllerRef, controllerKindSS.Kind, []string{"apps"})
	if !ok || err != nil {
		return nil, err
	}
	var ss appsv1.StatefulSet
	err = s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: controllerRef.Name}, &ss)
	if err != nil {
		// The only possible error is NotFound, which is ok here.
		return nil, nil
	}
	if ss.UID != controllerRef.UID {
		return nil, nil
	}

	return &controllerAndScale{ss.UID, *(ss.Spec.Replicas)}, nil
}

func (s *ScaleFinder) getPodDeployment(ctx context.Context, controllerRef *metav1.OwnerReference, namespace string) (*controllerAndScale, error) {
	ok, err := verifyGroupKind(controllerRef, controllerKindRS.Kind, []string{"apps", "extensions"})
	if !ok || err != nil {
		return nil, err
	}
	var rs appsv1.ReplicaSet
	err = s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: controllerRef.Name}, &rs)
	if err != nil {
		// The only possible error is NotFound, which is ok here.
		return nil, nil
	}
	if rs.UID != controllerRef.UID {
		return nil, nil
	}
	controllerRef = metav1.GetControllerOf(&rs)
	if controllerRef == nil {
		return nil, nil
	}

	ok, err = verifyGroupKind(controllerRef, controllerKindDep.Kind, []string{"apps", "extensions"})
	if !ok || err != nil {
		return nil, err
	}
	var deployment appsv1.Deployment
	err = s.client.Get(ctx, types.NamespacedName{Namespace: rs.Namespace, Name: controllerRef.Name}, &deployment)
	if err != nil {
		// The only possible error is NotFound, which is ok here.
		return nil, nil
	}
	if deployment.UID != controllerRef.UID {
		return nil, nil
	}
	return &controllerAndScale{deployment.UID, *(deployment.Spec.Replicas)}, nil
}

func (s *ScaleFinder) getPodReplicationController(ctx context.Context, controllerRef *metav1.OwnerReference, namespace string) (*controllerAndScale, error) {
	ok, err := verifyGroupKind(controllerRef, controllerKindRC.Kind, []string{""})
	if !ok || err != nil {
		return nil, err
	}
	var rc corev1.ReplicationController
	err = s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: controllerRef.Name}, &rc)
	if err != nil {
		// The only possible error is NotFound, which is ok here.
		return nil, nil
	}
	if rc.UID != controllerRef.UID {
		return nil, nil
	}
	return &controllerAndScale{rc.UID, *(rc.Spec.Replicas)}, nil
}

func (s *ScaleFinder) getScaleController(ctx context.Context, controllerRef *metav1.OwnerReference, namespace string) (*controllerAndScale, error) {
	gv, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		return nil, err
	}
	res := unstructured.Unstructured{}
	res.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    controllerRef.Kind,
	})
	res.SetNamespace(namespace)
	res.SetName(controllerRef.Name)
	scale := unstructured.Unstructured{}
	scale.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   autoscalingv1.SchemeGroupVersion.Group,
		Version: autoscalingv1.SchemeGroupVersion.Version,
		Kind:    "Scale",
	})
	err = s.client.SubResource("scale").Get(ctx, &res, &scale)
	if err != nil {
		if errors.IsNotFound(err) {
			// The IsNotFound error can mean either that the resource does not exist,
			// or it exist but doesn't implement the scale subresource. We check which
			// situation we are facing so we can give an appropriate error message.
			isScale, err := s.implementsScale(gv, controllerRef.Kind)
			if err != nil {
				return nil, err
			}
			if !isScale {
				return nil, fmt.Errorf("%s does not implement the scale subresource", res.GetAPIVersion())
			}
			return nil, nil
		}
		return nil, fmt.Errorf("cannot get scale subresource of %s: %w", res.GroupVersionKind().String(), err)
	}

	if scale.GetUID() != controllerRef.UID {
		return nil, nil
	}
	replicas, found, err := unstructured.NestedInt64(scale.Object, "spec", "replicas")
	if err != nil || !found {
		return nil, fmt.Errorf("unable to find spec.replicas in scale subresource %v", scale.Object)
	}
	return &controllerAndScale{scale.GetUID(), int32(replicas)}, nil
}

func (s *ScaleFinder) implementsScale(gv schema.GroupVersion, kind string) (bool, error) {
	resourceList, err := s.discoveryClient.ServerResourcesForGroupVersion(gv.String())
	if err != nil {
		return false, err
	}
	for _, resource := range resourceList.APIResources {
		if resource.Kind != kind {
			continue
		}
		if strings.HasSuffix(resource.Name, "/scale") {
			return true, nil
		}
	}
	return false, nil
}

func verifyGroupKind(controllerRef *metav1.OwnerReference, expectedKind string, expectedGroups []string) (bool, error) {
	gv, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		return false, err
	}

	if controllerRef.Kind != expectedKind {
		return false, nil
	}

	for _, group := range expectedGroups {
		if group == gv.Group {
			return true, nil
		}
	}

	return false, nil
}
