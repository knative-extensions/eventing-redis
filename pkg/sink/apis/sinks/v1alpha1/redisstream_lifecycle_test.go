/*
Copyright 2018 The Knative Authors

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

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

var (
	num            = int32(1)
	knativeservice = &servingv1.Service{
		Status: servingv1.ServiceStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{{
					Type:   "Ready",
					Status: "True",
				}},
			},
			RouteStatusFields: servingv1.RouteStatusFields{
				Address: &duckv1.Addressable{
					URL: apis.HTTP("example"),
				},
			},
		},
	}
)

var _ = duck.VerifyType(&RedisStreamSink{}, &duckv1.Conditions{})

func TestRedisStreamSinkGetConditionSet(t *testing.T) {
	r := &RedisStreamSink{}

	if got, want := r.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetTopLevelCondition=%v, want=%v", got, want)
	}
}

func TestRedisStreamSinkStatusIsReady(t *testing.T) {
	tests := []struct {
		name string
		s    *RedisStreamSinkStatus
		want bool
	}{{
		name: "uninitialized",
		s:    &RedisStreamSinkStatus{},
		want: false,
	}, {
		name: "initialized",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			return s
		}(),
		want: false,
	}, {
		name: "mark rolebinding",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			s.MarkRoleBinding()
			return s
		}(),
		want: false,
	}, {
		name: "mark serviceaccount",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			s.MarkServiceAccount()
			return s
		}(),
		want: false,
	}, {
		name: "mark deployed",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			s.PropagateKnativeServiceAddress(knativeservice)
			return s
		}(),
		want: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.s.IsReady()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("%s: unexpected condition (-want, +got) = %v", test.name, diff)
			}
		})
	}
}

func TestRedisStreamSinkStatusGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		s         *RedisStreamSinkStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name:      "uninitialized",
		s:         &RedisStreamSinkStatus{},
		condQuery: RedisStreamConditionReady,
		want:      nil,
	}, {
		name: "initialized",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:   RedisStreamConditionReady,
			Status: corev1.ConditionUnknown,
		},
	}, {
		name: "mark rolebinding and deployed",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			s.MarkRoleBinding()
			s.MarkKnativeService()
			s.PropagateKnativeServiceAddress(knativeservice)
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:   RedisStreamConditionReady,
			Status: corev1.ConditionTrue,
		},
	}, {
		name: "mark knativeService, rolebinding, then no knativeService",
		s: func() *RedisStreamSinkStatus {
			s := &RedisStreamSinkStatus{}
			s.InitializeConditions()
			s.MarkKnativeService()
			s.MarkRoleBinding()
			s.MarkNoKnativeService("Testing")
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:   RedisStreamConditionReady,
			Status: corev1.ConditionUnknown,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.s.GetCondition(test.condQuery)
			ignoreTime := cmpopts.IgnoreFields(apis.Condition{},
				"LastTransitionTime", "Severity")
			if diff := cmp.Diff(test.want, got, ignoreTime); diff != "" {
				t.Error("unexpected condition (-want, +got) =", diff)
			}
		})
	}
}
