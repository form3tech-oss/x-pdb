package lock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	coordv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = coordv1.AddToScheme(scheme)
}

func TestService_Lock(t *testing.T) {
	leaseNamespace := "default"
	leaseIdentity := "x-pdb-123"
	otherLeaseIdentity := "x-pdb-666-from-other-cluster"
	testPodSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "test",
		},
	}

	type args struct {
		podNamespace string
		podSelector  *metav1.LabelSelector
	}
	tests := []struct {
		name          string
		existingLease *coordv1.Lease
		args          args
		wantErr       bool
		assert        func(client.Client)
	}{
		{
			name: "should acquire lock",
			args: args{
				podNamespace: "default",
				podSelector:  testPodSelector,
			},
			wantErr: false,
			assert: func(cl client.Client) {
				expectedLease := createLeaseForSelector(leaseNamespace, leaseIdentity, "default", testPodSelector)
				err := cl.Get(context.Background(), client.ObjectKeyFromObject(expectedLease), expectedLease)
				assert.NoError(t, err, "get lease failed")
				assert.Equal(t, leaseIdentity, *expectedLease.Spec.HolderIdentity)
			},
		},
		{
			name:          "should error if lock already exists and has not expired",
			existingLease: makeTestLease(leaseNamespace, otherLeaseIdentity, "default", testPodSelector),
			args: args{
				podNamespace: "default",
				podSelector:  testPodSelector,
			},
			wantErr: true,
			assert: func(cl client.Client) {
				expectedLease := createLeaseForSelector(leaseNamespace, leaseIdentity, "default", testPodSelector)
				err := cl.Get(context.Background(), client.ObjectKeyFromObject(expectedLease), expectedLease)
				assert.NoError(t, err, "get lease failed")
				assert.NotEqual(t, *expectedLease.Spec.HolderIdentity, leaseIdentity)
			},
		},
		{
			name:          "should take over if lock already exists and has expired",
			existingLease: makeTestLease(leaseNamespace, otherLeaseIdentity, "default", testPodSelector, leaseExpired),
			args: args{
				podNamespace: "default",
				podSelector:  testPodSelector,
			},
			wantErr: false,
			assert: func(cl client.Client) {
				// verify that we took over the lease by verifying the identity
				expectedLease := createLeaseForSelector(leaseNamespace, leaseIdentity, "default", testPodSelector)
				err := cl.Get(context.Background(), client.ObjectKeyFromObject(expectedLease), expectedLease)
				assert.NoError(t, err, "get lease failed")
				assert.Equal(t, *expectedLease.Spec.HolderIdentity, leaseIdentity)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)
			if tt.existingLease != nil {
				clientBuilder.WithObjects(tt.existingLease)
			}
			cl := clientBuilder.Build()
			logger := zap.New(zap.UseDevMode(true))
			s := NewService(&logger, cl, cl, nil, leaseNamespace, nil)
			if err := s.LocalLock(context.Background(), leaseIdentity, tt.args.podNamespace, tt.args.podSelector); (err != nil) != tt.wantErr {
				t.Errorf("Service.LockXPDB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.assert != nil {
				tt.assert(cl)
			}
		})
	}
}

type testLeaseFunc func(*coordv1.Lease)

func makeTestLease(leaseNs, leaseID, podNs string, podSelector *metav1.LabelSelector, modifier ...testLeaseFunc) *coordv1.Lease {
	lease := createLeaseForSelector(leaseNs, leaseID, podNs, podSelector)
	for _, m := range modifier {
		m(lease)
	}
	return lease
}

func leaseExpired(lease *coordv1.Lease) {
	lease.Spec.AcquireTime = &metav1.MicroTime{Time: time.Now().Add(-time.Minute)}
}
