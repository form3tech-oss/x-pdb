package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/form3tech-oss/x-pdb/api/v1alpha1"
	"github.com/form3tech-oss/x-pdb/internal/pdb"
	"github.com/form3tech-oss/x-pdb/internal/preactivities"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	coordv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const clusterContext = "kind-x-pdb-%d"

var (
	scheme        = runtime.NewScheme()
	waitPeriod    = time.Minute * 3
	retryInterval = time.Second * 3
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(policyv1.AddToScheme(scheme))
}

type testStage struct {
	ctx  context.Context
	st   *suite.Suite
	name string
}

func newScenario(s *multiClusterTestSuite) (*testStage, *testStage, *testStage) {
	stage := &testStage{
		ctx:  context.Background(),
		st:   &s.Suite,
		name: fmt.Sprintf("multi-cluster-suite-%d", time.Now().UnixNano()),
	}
	s.Suite.T().Cleanup(stage.cleanup)
	return stage, stage, stage
}

type ClusterContext struct {
	clusterID int
	testStage *testStage
	client    client.Client
	cs        *kubernetes.Clientset
}

type Modifier func(*ClusterContext)

func (s *testStage) cleanup() {
	if !s.st.T().Failed() {
		return
	}

	// dump logs from x-pdb pods
	// only if tests have failed
	for i := 1; i <= 3; i++ {
		cc, err := makeClusterContext(s, i)
		s.st.NoError(err, "unable to create cluster context")
		var podList v1.PodList
		err = cc.client.List(s.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
			"app.kubernetes.io/name": "x-pdb",
		})
		s.st.NoError(err, "unable to list pods")
		for i := range podList.Items {
			po := podList.Items[i]
			logs, err := getPodLogs(cc, po.Name)
			s.st.NoError(err, "unable to get pod logs")
			s.st.T().Logf("pod %s logs: %s", po.Name, logs)
		}
	}
}

func getPodLogs(cc *ClusterContext, podName string) (string, error) {
	req := cc.cs.CoreV1().Pods("default").GetLogs(podName, &v1.PodLogOptions{
		TailLines: ptr.To(int64(1000)),
	})
	podLogs, err := req.Stream(cc.testStage.ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, podLogs)
	if err != nil {
		return "", err
	}
	str := buf.String()
	return str, nil
}

func (s *testStage) in_cluster_1(modifier ...Modifier) *testStage {
	return s.in_cluster(1, modifier...)
}

func (s *testStage) in_all_clusters(modifier ...Modifier) *testStage {
	for i := 1; i <= 3; i++ {
		s.in_cluster(i, modifier...)
	}
	return s
}

func (s *testStage) in_cluster(clusterID int, modifier ...Modifier) *testStage {
	clusterContext, err := makeClusterContext(s, clusterID)
	s.st.NoError(err, "unable to create cluster context")
	for _, m := range modifier {
		m(clusterContext)
	}
	return s
}

func cleanup_leases_across_all_clusters(cc *ClusterContext) {
	cc.testStage.st.T().Log("cleaning up dangling leases in all clusters")
	for i := 1; i <= 3; i++ {
		cx, err := makeClusterContext(cc.testStage, i)
		cc.testStage.st.NoError(err, "unable to create cluster context")
		var lease coordv1.Lease
		err = cx.client.DeleteAllOf(cc.testStage.ctx, &lease, client.InNamespace("kube-system"), client.MatchingLabels{
			"app": "x-pdb",
		})
		cc.testStage.st.NoError(err, "unable clean up dangling leases")
	}
}

func wait_until_all_pods_are_ready(cc *ClusterContext) {
	err := wait.PollUntilContextTimeout(context.Background(), retryInterval, waitPeriod, true, func(ctx context.Context) (done bool, err error) {
		var podList v1.PodList
		if err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
			"app": "test",
		}); err != nil {
			return false, err
		}
		var deployment appsv1.Deployment
		if err := cc.client.Get(cc.testStage.ctx, types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		}, &deployment); err != nil {
			return false, err
		}
		if countHealthyPods(podList.Items) >= *deployment.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})
	cc.testStage.st.NoError(err, "not all pods are Ready")
}

func wait_5_seconds(_ *ClusterContext) {
	<-time.After(time.Second * 5)
}

func make_all_pods_ready(cc *ClusterContext) {
	var podList v1.PodList
	var deployment appsv1.Deployment
	// wait for the exact number of pods to be created
	err := wait.PollUntilContextTimeout(context.Background(), retryInterval, waitPeriod, true, func(ctx context.Context) (done bool, err error) {
		if err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
			"app": "test",
		}); err != nil {
			return false, err
		}
		if err := cc.client.Get(cc.testStage.ctx, types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		}, &deployment); err != nil {
			return false, err
		}
		if len(podList.Items) >= int(*deployment.Spec.Replicas) {
			return true, nil
		}
		return false, nil
	})
	cc.testStage.st.NoError(err, "not all pods are Ready")

	configMapData := make(map[string]string)
	for i := range podList.Items {
		configMapData[podList.Items[i].Name] = "true"
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
	}
	_, err = controllerutil.CreateOrUpdate(cc.testStage.ctx, cc.client, cm, func() error {
		cm.Data = configMapData
		return nil
	})
	cc.testStage.st.T().Cleanup(func() {
		_ = cc.client.Delete(cc.testStage.ctx, cm)
	})
	cc.testStage.st.NoError(err, "unable to create/update configmap")

	wait_until_all_pods_are_ready(cc)
}

func make_all_pods_always_ready(cc *ClusterContext) {
	configMapData := map[string]string{
		"always": "true",
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
	}
	_, err := controllerutil.CreateOrUpdate(cc.testStage.ctx, cc.client, cm, func() error {
		cm.Data = configMapData
		return nil
	})
	cc.testStage.st.T().Cleanup(func() {
		_ = cc.client.Delete(cc.testStage.ctx, cm)
	})
	cc.testStage.st.NoError(err, "unable to create/update configmap")

	wait_until_all_pods_are_ready(cc)
}

func make_one_pod_unready(cc *ClusterContext) {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")
	configMapData := make(map[string]string)
	var unreadyPod string
	for i := range podList.Items {
		if i == 0 {
			unreadyPod = podList.Items[i].Name
			configMapData[podList.Items[i].Name] = "false"
		} else {
			configMapData[podList.Items[i].Name] = "true"
		}
	}

	cm := newConfigMap()
	cm.Data = configMapData
	err = cc.client.Update(cc.testStage.ctx, cm)
	cc.testStage.st.NoError(err, "unable to update configmap")

	// wait for it to be unready
	err = wait.PollUntilContextTimeout(context.Background(), retryInterval, waitPeriod, true, func(ctx context.Context) (done bool, err error) {
		var pod v1.Pod
		if err := cc.client.Get(cc.testStage.ctx, types.NamespacedName{
			Name:      unreadyPod,
			Namespace: "default",
		}, &pod); err != nil {
			return false, err
		}
		return !pdb.IsPodReady(&pod), nil
	})
	cc.testStage.st.NoError(err, "not all pods are Ready")
}

func newConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
	}
}

func evicting_one_pod_must_not_be_allowed(cc *ClusterContext) {
	err := try_evict_pod(cc)
	cc.testStage.st.ErrorContains(err, "Cannot disrupt pod as it would violate the pod's xpdb disruption budget.", "unexpected error trying to evict pod")
}

func evicting_one_pod_must_not_be_allowed_by_disruption_probe(cc *ClusterContext) {
	err := try_evict_pod(cc)
	cc.testStage.st.ErrorContains(err, "Cannot disrupt pod as the pod's xpdb disruption probe didn't allow it.")
}

func evicting_one_pod_must_not_be_allowed_by_disruption_preactivities(cc *ClusterContext) {
	err := try_evict_pod(cc)
	cc.testStage.st.ErrorContains(err, "Cannot disrupt pod has it has pending disruption pre-activities.")
}

func evicting_one_pod_must_be_allowed(cc *ClusterContext) {
	err := evict_pod_until_successful(cc)
	cc.testStage.st.NoError(err, "eviction was not possible in %d", cc.clusterID)
}

func evicting_all_pods_must_be_allowed(cc *ClusterContext) {
	err := try_evict_all_pods(cc)
	cc.testStage.st.NoError(err, "eviction was not possible in %d", cc.clusterID)
}

func try_evict_all_pods(cc *ClusterContext) error {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")
	for i := range podList.Items {
		err = cc.client.SubResource("eviction").Create(cc.testStage.ctx, &podList.Items[i], &policyv1.Eviction{})
		cc.testStage.st.NoError(err, "unable to evict pod")
	}
	return nil
}

func evict_pod_until_successful(cc *ClusterContext) error {
	var evictionErr error
	err := wait.PollUntilContextTimeout(context.Background(), retryInterval, waitPeriod, true, func(ctx context.Context) (done bool, err error) {
		evictionErr = try_evict_pod(cc)
		if evictionErr == nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("eviction was not possible: %w", evictionErr)
	}
	return nil
}

func try_evict_pod(cc *ClusterContext) error {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")

	for i := range podList.Items {
		if !pdb.IsPodReady(&podList.Items[i]) {
			continue
		}
		// take the first ready pod and try to evict it
		return cc.client.SubResource("eviction").Create(cc.testStage.ctx, &podList.Items[i], &policyv1.Eviction{})
	}
	return fmt.Errorf("no ready pods found which we can evict")
}

func countHealthyPods(pods []v1.Pod) (currentHealthy int32) {
	for i := range pods {
		// Pod is being deleted.
		if pods[i].DeletionTimestamp != nil {
			continue
		}
		// Pod is expected to be deleted soon.
		if pdb.IsPodReady(&pods[i]) {
			currentHealthy++
		}
	}
	return
}

const testDeploymentName = "test"

func a_deployment_with_three_replicas(cc *ClusterContext) {
	// We do delete the deployment/replicasets below,
	// however in CI it happens that orphaned pods stick around
	// and are not cleaned up within a reasonable time frame.
	// This lets tests fail, because the "controller" field is empty (rs has been deleted)
	// and xpdb is not able to figure out the "desired" / "expected" count
	// see: internal/pdb/scale.go
	cc.testStage.st.T().Cleanup(func() {
		cc.testStage.st.T().Logf("deleting all pods with label app=test")
		var po v1.Pod
		err := cc.client.DeleteAllOf(cc.testStage.ctx, &po, client.InNamespace("default"), client.MatchingLabels{
			"app": "test",
		})
		cc.testStage.st.NoError(err, "unable to clean up pod")

		err = wait.PollUntilContextTimeout(context.Background(), retryInterval, waitPeriod, true, func(ctx context.Context) (done bool, err error) {
			cc.testStage.st.T().Logf("waiting until all pods with label app=test are gone")
			var podList v1.PodList
			if err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
				"app": "test",
			}); err != nil {
				return false, err
			}
			if len(podList.Items) == 0 {
				return true, nil
			}
			return false, nil
		})
		cc.testStage.st.NoError(err, "error waiting for pod removal")
	})
	create_resource_with_cleanup(cc, newDeployment(3))
	create_resource_with_cleanup(cc, newServiceAccount())
	create_resource_with_cleanup(cc, newRole())
	create_resource_with_cleanup(cc, newRoleBinding())
}

func create_resource_with_cleanup(cc *ClusterContext, obj client.Object) {
	err := cc.client.Create(cc.testStage.ctx, obj)
	cc.testStage.st.NoError(err, "failed to create resource")
	cc.testStage.st.T().Cleanup(func() {
		cc.testStage.st.T().Logf("deleting resource %T %s", obj, obj.GetName())
		err = cc.client.Delete(cc.testStage.ctx, obj)
		cc.testStage.st.NoError(err, "unable to clean up resource")
	})
}

func newDeployment(replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testDeploymentName,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName:            "test",
					AutomountServiceAccountToken:  ptr.To(true),
					TerminationGracePeriodSeconds: ptr.To(int64(0)),
					Containers: []v1.Container{
						{
							Name:            "test",
							Image:           "x-pdb-test:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name: "NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							ReadinessProbe: &v1.Probe{
								SuccessThreshold: 1,
								FailureThreshold: 1,
								PeriodSeconds:    1,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Port: intstr.FromInt(8080),
										Path: "/",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newServiceAccount() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
}

func newRole() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs: []string{
					"get", "list", "watch",
				},
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
				},
			},
		},
	}
}

func newRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "test",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test",
				Namespace: "default",
			},
		},
	}
}

func makeClusterContext(stage *testStage, clusterID int) (*ClusterContext, error) {
	kubeContext := fmt.Sprintf(clusterContext, clusterID)
	cfg, err := config.GetConfigWithContext(kubeContext)
	if err != nil {
		return nil, err
	}
	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Cache: &client.CacheOptions{
			DisableFor: []client.Object{
				&v1.Pod{},
				&v1.ConfigMap{},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &ClusterContext{
		clusterID: clusterID,
		testStage: stage,
		client:    client,
		cs:        cs,
	}, nil
}

func a_pdb_with_min_available_2(cc *ClusterContext) {
	create_resource_with_cleanup(cc, newPDB(ptr.To(int32(2)), nil))
}

func a_pdb_with_max_unavailable_1(cc *ClusterContext) {
	create_resource_with_cleanup(cc, newPDB(nil, ptr.To(int32(1))))
}

func newPDB(minAvailable, maxUnavailable *int32) *policyv1.PodDisruptionBudget {
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
		},
	}
	if minAvailable != nil {
		pdb.Spec.MinAvailable = &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: *minAvailable,
		}
	}
	if maxUnavailable != nil {
		pdb.Spec.MaxUnavailable = &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: *maxUnavailable,
		}
	}
	return pdb
}

func a_xpdb_with_min_available_8(cc *ClusterContext) {
	create_resource_with_cleanup(cc, newXPDB(ptr.To(int32(8)), nil, nil, nil))
}

func a_xpdb_with_max_unavailable_1(cc *ClusterContext) {
	create_resource_with_cleanup(cc, newXPDB(nil, ptr.To(int32(1)), nil, nil))
}

func a_suspended_xpdb_with_max_unavailable_1(cc *ClusterContext) {
	create_resource_with_cleanup(cc, newXPDB(nil, ptr.To(int32(1)), ptr.To(true), nil))
}

func a_xpdb_with_max_unavailable_1_and_probe(cc *ClusterContext) {
	probe := &v1alpha1.XPodDisruptionBudgetProbeSpec{
		Endpoint: "test-disruption-probe.default.svc.cluster.local:8080",
	}
	create_resource_with_cleanup(cc, newXPDB(nil, ptr.To(int32(1)), nil, probe))
}

func newXPDB(minAvailable, maxUnavailable *int32, suspended *bool, probe *v1alpha1.XPodDisruptionBudgetProbeSpec) *v1alpha1.XPodDisruptionBudget {
	xpdb := &v1alpha1.XPodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.XPodDisruptionBudgetSpec{
			Suspend: suspended,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
		},
	}
	if minAvailable != nil {
		xpdb.Spec.MinAvailable = &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: *minAvailable,
		}
	}
	if maxUnavailable != nil {
		xpdb.Spec.MaxUnavailable = &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: *maxUnavailable,
		}
	}
	if probe != nil {
		xpdb.Spec.Probe = probe
	}
	return xpdb
}

const for_2_minutes = time.Minute * 2

func (s *testStage) run_race_test(testDuration time.Duration, testFunc func(cc *ClusterContext), verifyFunc func(ccs []*ClusterContext)) *testStage {
	timeoutCtx, cancel := context.WithTimeout(s.ctx, testDuration)
	defer cancel()

	var allClusterCtx []*ClusterContext
	for i := 1; i <= 3; i++ {
		cc, err := makeClusterContext(s, i)
		s.st.NoError(err, "unable to create cluster context")
		allClusterCtx = append(allClusterCtx, cc)
		go run_until_done(timeoutCtx, func() {
			testFunc(cc)
		})
	}

	go func() {
		for {
			select {
			case <-timeoutCtx.Done():
				return
			default:
				verifyFunc(allClusterCtx)
				<-time.After(time.Millisecond * 25)
			}
		}
	}()
	s.st.T().Log("wait for race run completion")
	<-timeoutCtx.Done()
	s.st.T().Log("done with race run")

	return s
}

func run_until_done(ctx context.Context, runFunc func()) {
	minJitter := 40
	maxJitter := 400
	//nolint:gosec
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			runFunc()
			jitter := rng.Intn(maxJitter-minJitter) + minJitter
			<-time.After(time.Duration(jitter * int(time.Millisecond)))
		}
	}
}

func evict_all_pods(cc *ClusterContext) {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.InNamespace("default"), client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")
	for i := range podList.Items {
		err = cc.client.SubResource("eviction").Create(cc.testStage.ctx, &podList.Items[i], &policyv1.Eviction{})
		if err == nil {
			cc.testStage.st.T().Logf("[%d] %s evicted %s", cc.clusterID, time.Now().String(), podList.Items[i].Name)
		}
	}
}

func ensure_pods_max_unavailable_1(clusters []*ClusterContext) {
	var expectedPods int
	var readyPods int
	for _, cc := range clusters {
		var deployment appsv1.Deployment
		err := cc.client.Get(cc.testStage.ctx, types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		}, &deployment)
		cc.testStage.st.NoError(err, "unable to list pods")
		expectedPods += int(*deployment.Spec.Replicas)

		var podList v1.PodList
		err = cc.client.List(cc.testStage.ctx, &podList, client.MatchingLabels{
			"app": "test",
		})
		cc.testStage.st.NoError(err, "unable to list pods")
		for i := range podList.Items {
			if pdb.IsPodReady(&podList.Items[i]) {
				readyPods++
			}
		}
	}

	if expectedPods-readyPods > 1 {
		clusters[0].testStage.st.FailNow("max unavailable breached, aborting test immediately")
	}
}

func make_pods_disruptable_on_disruption_probe(cc *ClusterContext) {
	var configmap v1.ConfigMap
	err := cc.client.Get(cc.testStage.ctx, types.NamespacedName{
		Name:      "test-disruption-probe-config",
		Namespace: "default",
	}, &configmap)
	cc.testStage.st.NoError(err, "unable to get test disruption probe config")

	configmap.Data["notAllowedDisruptions"] = ""

	err = cc.client.Update(cc.testStage.ctx, &configmap)
	cc.testStage.st.NoError(err, "could not update disruption probe configmap")
}

func make_pods_not_disruptable_on_disruption_probe(cc *ClusterContext) {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")

	notAllowedDisruptionsList := []string{}
	for _, p := range podList.Items {
		notAllowedDisruptionsList = append(notAllowedDisruptionsList, fmt.Sprintf("%s/%s", p.Namespace, p.Name))
	}

	var configmap v1.ConfigMap
	err = cc.client.Get(cc.testStage.ctx, types.NamespacedName{
		Name:      "test-disruption-probe-config",
		Namespace: "default",
	}, &configmap)
	cc.testStage.st.NoError(err, "unable to get test disruption probe config")

	configmap.Data["notAllowedDisruptions"] = strings.Join(notAllowedDisruptionsList, ",")

	err = cc.client.Update(cc.testStage.ctx, &configmap)
	cc.testStage.st.NoError(err, "could not update disruption probe configmap")

	// Give some time for probe to pick up the config change
	time.Sleep(30 * time.Second)
}

func test_disruption_probe_is_installed(cc *ClusterContext) {
	//nolint:gosec
	cmd := exec.Command("helm", "upgrade", "-i", "test-disruption-probe", "./hack/env/charts/test-disruption-probe",
		"--namespace", "default",
		"-f", "./tests/resources/test-disruption-probe-values.yaml",
		"--kube-context", fmt.Sprintf("kind-x-pdb-%d", cc.clusterID))

	_, err := runCmd(cmd)
	cc.testStage.st.NoError(err, "unable to install test-disruption-probe")

	cc.testStage.st.T().Cleanup(func() {
		//nolint:gosec
		cmd := exec.Command("helm", "uninstall", "test-disruption-probe",
			"--namespace", "default",
			"--kube-context", fmt.Sprintf("kind-x-pdb-%d", cc.clusterID))

		_, err := runCmd(cmd)
		cc.testStage.st.NoError(err, "unable to install test-disruption-probe")
	})
}

func pods_have_preactivities(cc *ClusterContext) {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")

	for _, pod := range podList.Items {
		if pod.Annotations == nil {
			pod.Annotations = map[string]string{}
		}

		pod.Annotations[preactivities.PreActivityAnnotationNamePrefix+"test"] = "true"
		err := cc.client.Update(cc.testStage.ctx, &pod)
		cc.testStage.st.NoError(err, "unable to update pod")
	}
}

func pods_preactivities_are_deleted(cc *ClusterContext) {
	var podList v1.PodList
	err := cc.client.List(cc.testStage.ctx, &podList, client.MatchingLabels{
		"app": "test",
	})
	cc.testStage.st.NoError(err, "unable to list pods")

	for _, pod := range podList.Items {
		if pod.Annotations != nil {
			delete(pod.Annotations, preactivities.PreActivityAnnotationNamePrefix+"test")
		}
		err := cc.client.Update(cc.testStage.ctx, &pod)
		cc.testStage.st.NoError(err, "unable to update pod")
	}
}
