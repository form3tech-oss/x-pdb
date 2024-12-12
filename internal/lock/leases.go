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
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	coordv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	leaseAnnotationNamespace = "xpdb.form3.tech/pod-namespace"
	leaseAnnotationSelector  = "xpdb.form3.tech/pod-selector"
)

// LeaseDurationSeconds is the default duration for the Lease object.
// This defines the duration for how long a given xpdb is locked.
// The value should be higher than the sum of the following durations:
//   - the round-trip latency across all clusters (1-2 seconds)
//   - the processing time of the x-pdb http server (<1 second)
//   - the time kube-apiserver needs to process the x-pdb admission control response
//     and the time it takes until the desired action (evict/delete pod) is observable through the kube-apiserver (1-2 seconds)
//   - a generous surcharge (1-... seconds)
var LeaseDurationSeconds = int32(5)

// CreateLeaseHolderIdentity creates the leaseHolderIdentity for the lease used to perform
// the locking mechanism of the xpdb disruptions.
func CreateLeaseHolderIdentity(clusterID, xpdbPodID, podNamespace, podName string) string {
	// The lock that we acquire is unique per:
	// {cluster}/{x-pdb-pod}/{target-pod-namespace}/{target-pod-name}/{admission-request-uuid}
	// This way we ensure that we are not able to evict a different pod matching the same xpdb
	// or evicting the same pod twice while it's being processed.
	return fmt.Sprintf("%s/%s/%s/%s/%s", clusterID, xpdbPodID, podNamespace, podName, uuid.New().String())
}

func createLeaseNameForSelector(namespace string, selector *metav1.LabelSelector) string {
	leaseHash := sha256.New()
	leaseHash.Write([]byte(namespace))
	leaseHash.Write([]byte(selector.String()))
	leaseHashBytes := leaseHash.Sum(nil)
	// using base32 to prevent usage of special characters like + and /
	// trim padding character `=`
	prefixStr := strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(leaseHashBytes[0:24]), "="))
	return fmt.Sprintf("xpdb-%s", prefixStr)
}

func createLeaseForSelector(leaseNamespace, leaseHolderIdentity, namespace string, selector *metav1.LabelSelector) *coordv1.Lease {
	return &coordv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createLeaseNameForSelector(namespace, selector),
			Namespace: leaseNamespace,
			Annotations: map[string]string{
				leaseAnnotationNamespace: namespace,
				leaseAnnotationSelector:  selector.String(),
			},
			Labels: map[string]string{
				"app": "x-pdb",
			},
		},
		Spec: coordv1.LeaseSpec{
			HolderIdentity:       ptr.To(leaseHolderIdentity),
			AcquireTime:          &metav1.MicroTime{Time: time.Now()},
			LeaseDurationSeconds: &LeaseDurationSeconds,
		},
	}
}
