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
	"testing"

	xpdbv1alpha1 "github.com/form3tech-oss/x-pdb/api/v1alpha1"
	coordv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = coordv1.AddToScheme(scheme)
}

func TestService_disruptionAllowed(t *testing.T) {
	type args struct {
		xpdb          *xpdbv1alpha1.XPodDisruptionBudget
		pod           *corev1.Pod
		expectedCount int32
		healthyCount  int32
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "disruption allowed if no PDB passed in",
			args:    args{},
			want:    true,
			wantErr: false,
		},
		{
			name: "disruption allowed if with maxUnavailable int",
			args: args{
				expectedCount: 3,
				healthyCount:  3,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MaxUnavailable: ptr.To(intstr.FromInt(1)),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "disruption not allowed if with maxUnavailable int",
			args: args{
				expectedCount: 3,
				healthyCount:  2,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MaxUnavailable: ptr.To(intstr.FromInt(1)),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "disruption allowed if with maxUnavailable is 1 and pod is not ready",
			args: args{
				expectedCount: 3,
				healthyCount:  2,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MaxUnavailable: ptr.To(intstr.FromInt(1)),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "disruption allowed if with maxUnavailable percent",
			args: args{
				expectedCount: 100,
				healthyCount:  95,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MaxUnavailable: ptr.To(intstr.FromString("90%")),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "disruption not allowed if with maxUnavailable percent",
			args: args{
				expectedCount: 100,
				healthyCount:  90,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MaxUnavailable: ptr.To(intstr.FromString("10%")),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "disruption allowed if with minAvailable int",
			args: args{
				expectedCount: 3,
				healthyCount:  3,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MinAvailable: ptr.To(intstr.FromInt(2)),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "disruption not allowed if with minAvailable int",
			args: args{
				expectedCount: 3,
				healthyCount:  2,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MinAvailable: ptr.To(intstr.FromInt(2)),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "disruption allowed if with minAvailable percent",
			args: args{
				expectedCount: 100,
				healthyCount:  100,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MinAvailable: ptr.To(intstr.FromString("90%")),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "disruption not allowed if with minAvailable percent",
			args: args{
				expectedCount: 100,
				healthyCount:  90,
				xpdb: &xpdbv1alpha1.XPodDisruptionBudget{
					Spec: xpdbv1alpha1.XPodDisruptionBudgetSpec{
						MinAvailable: ptr.To(intstr.FromString("90%")),
					},
				},
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				logger: zap.New(),
			}
			got, err := s.disruptionAllowed(tt.args.xpdb, tt.args.pod, tt.args.expectedCount, tt.args.healthyCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.disruptionAllowed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Service.disruptionAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
