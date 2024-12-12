package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestMultiClusterTestSuite(t *testing.T) {
	log.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))
	suite.Run(t, &multiClusterTestSuite{})
}

type multiClusterTestSuite struct {
	suite.Suite
	name string
}

func (s *multiClusterTestSuite) SetupTest() {
	s.name = fmt.Sprintf("multi-cluster-test-%d", time.Now().UnixNano())
}

func (s *multiClusterTestSuite) TestMinAvailableDisruptionBlocking() {
	given, when, then := newScenario(s)

	given.
		in_all_clusters(
			// 9 pods total, 3 per cluster
			// 2 minAvailable per cluster (1 disruption per cluster allowed)
			// 8 minAvailable total (1 disruption allowed across clusters)
			a_deployment_with_three_replicas,
			a_pdb_with_min_available_2,
			a_xpdb_with_min_available_8,
		)

	// cluster 1 => should have a disruption
	// attempt eviction in all clusters
	when.
		in_all_clusters(
			make_all_pods_ready,
		).
		in_cluster_1(
			make_one_pod_unready,
		)
	then.
		in_all_clusters(
			evicting_one_pod_must_not_be_allowed,
			cleanup_leases_across_all_clusters,
		)
}

func (s *multiClusterTestSuite) TestMinAvailableDisruption() {
	given, when, then := newScenario(s)

	given.
		in_all_clusters(
			// 9 pods total, 3 per cluster
			// 2 minAvailable per cluster (1 disruption per cluster allowed)
			// 8 minAvailable total (1 disruption allowed across clusters)
			a_deployment_with_three_replicas,
			a_pdb_with_min_available_2,
			a_xpdb_with_min_available_8,
		)

	// all pods across clusters are ready,
	// attempt eviction in all clusters
	when.
		in_all_clusters(
			make_all_pods_ready,
		)
	then.
		in_all_clusters(
			evicting_one_pod_must_be_allowed,
			// wait for eviction to happen an new pod to spawn
			// before we reconcile the configmap again
			wait_5_seconds,
			make_all_pods_ready,
		)
}

func (s *multiClusterTestSuite) TestMaxUnavailableDisruptionBlocking() {
	given, when, then := newScenario(s)

	given.
		in_all_clusters(
			// 9 pods total, 3 per cluster
			// 2 minAvailable per cluster (1 disruption per cluster allowed)
			// 8 minAvailable total (1 disruption allowed across clusters)
			a_deployment_with_three_replicas,
			a_pdb_with_max_unavailable_1,
			a_xpdb_with_max_unavailable_1,
		)

	// cluster 1 => should have a disruption
	// attempt eviction in all clusters
	when.
		in_all_clusters(
			make_all_pods_ready,
		).
		in_cluster_1(
			make_one_pod_unready,
		)
	then.
		in_all_clusters(
			evicting_one_pod_must_not_be_allowed,
			cleanup_leases_across_all_clusters,
		)
}

func (s *multiClusterTestSuite) TestMaxUnavailableDisruption() {
	given, when, then := newScenario(s)

	given.
		in_all_clusters(
			// 9 pods total, 3 per cluster
			// 2 minAvailable per cluster (1 disruption per cluster allowed)
			// 8 minAvailable total (1 disruption allowed across clusters)
			a_deployment_with_three_replicas,
			a_pdb_with_max_unavailable_1,
			a_xpdb_with_max_unavailable_1,
		)

	// all pods across clusters are ready,
	// attempt eviction in all clusters
	when.
		in_all_clusters(
			make_all_pods_ready,
		)
	then.
		in_all_clusters(
			evicting_one_pod_must_be_allowed,
			// wait for eviction to happen an new pod to spawn
			// before we reconcile the configmap again
			wait_5_seconds,
			make_all_pods_ready,
			cleanup_leases_across_all_clusters,
		)
}

func (s *multiClusterTestSuite) TestSuspendDisruption() {
	given, when, then := newScenario(s)

	given.
		in_all_clusters(
			// 9 pods total, 3 per cluster
			// 2 minAvailable per cluster (1 disruption per cluster allowed)
			// 8 minAvailable total (1 disruption allowed across clusters)
			a_deployment_with_three_replicas,
			a_suspended_xpdb_with_max_unavailable_1,
		)

	// all pods across clusters are ready,
	// attempt eviction in all clusters
	when.
		in_all_clusters(
			make_all_pods_ready,
		)
	then.
		in_all_clusters(
			evicting_all_pods_must_be_allowed,
			// wait for eviction to happen an new pod to spawn
			// before we reconcile the configmap again
			wait_5_seconds,
			make_all_pods_ready,
			cleanup_leases_across_all_clusters,
		)
}

func (s *multiClusterTestSuite) TestEvictionRace() {
	given, when, then := newScenario(s)

	given.
		in_all_clusters(
			// 9 pods total, 3 per cluster
			// Note: no PDB set! Without XPDB all pods
			a_deployment_with_three_replicas,
			a_xpdb_with_max_unavailable_1,
		)

	when.
		in_all_clusters(
			make_all_pods_always_ready,
		)

	then.
		run_race_test(
			for_2_minutes,
			evict_all_pods,
			ensure_pods_max_unavailable_1,
		).
		in_all_clusters(cleanup_leases_across_all_clusters)
}

func (s *multiClusterTestSuite) TestDisruptionProbeAllowDisruption() {
	given, when, then := newScenario(s)

	given.in_cluster_1(
		test_disruption_probe_is_installed,
		a_deployment_with_three_replicas,
		a_xpdb_with_max_unavailable_1_and_probe,
	)

	when.in_cluster_1(
		make_pods_disruptable_on_disruption_probe,
		make_all_pods_always_ready,
	)

	then.in_cluster_1(
		evicting_one_pod_must_be_allowed,
	)
}

func (s *multiClusterTestSuite) TestDisruptionProbeDisallowDisruption() {
	given, when, then := newScenario(s)

	given.in_cluster_1(
		test_disruption_probe_is_installed,
		a_deployment_with_three_replicas,
		a_xpdb_with_max_unavailable_1_and_probe,
	)

	when.in_cluster_1(
		make_pods_not_disruptable_on_disruption_probe,
		make_all_pods_always_ready,
	)

	then.in_cluster_1(
		evicting_one_pod_must_not_be_allowed_by_disruption_probe,
	)
}

func (s *multiClusterTestSuite) TestPendingActivities() {
	given, when, then := newScenario(s)

	given.in_cluster_1(
		a_deployment_with_three_replicas,
	)

	when.in_cluster_1(
		make_all_pods_always_ready,
		pods_have_preactivities,
	)

	then.in_cluster_1(
		evicting_one_pod_must_not_be_allowed_by_disruption_preactivities,
	)

	then.in_cluster_1(
		pods_preactivities_are_deleted,
		evicting_one_pod_must_be_allowed,
	)
}
