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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	num = int32(1)
	availableStatefulSet = &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &num,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas: 1,
		},
	}
)

var _ = duck.VerifyType(&RedisStreamSource{}, &duckv1.Conditions{})

func TestRedisStreamSourceGetConditionSet(t *testing.T) {
	r := &RedisStreamSource{}

	if got, want := r.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetTopLevelCondition=%v, want=%v", got, want)
	}
}

func TestRedisStreamSourceStatusIsReady(t *testing.T) {
	tests := []struct {
		name string
		s    *RedisStreamSourceStatus
		want bool
	}{{
		name: "uninitialized",
		s:    &RedisStreamSourceStatus{},
		want: false,
	}, {
		name: "initialized",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			return s
		}(),
		want: false,
	}, {
		name: "mark sink",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.MarkSink(apis.HTTP("example").String())
			return s
		}(),
		want: false,
	}, {
		name: "mark rolebinding",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.MarkRoleBinding()
			return s
		}(),
		want: false,
	}, {
		name: "mark serviceaccount",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.MarkServiceAccount()
			return s
		}(),
		want: false,
	}, {
		name: "mark deployed",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.PropagateStatefulSetAvailability(availableStatefulSet)
			return s
		}(),
		want: false,
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

func TestRedisStreamSourceStatusGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		s         *RedisStreamSourceStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name:      "uninitialized",
		s:         &RedisStreamSourceStatus{},
		condQuery: RedisStreamConditionReady,
		want:      nil,
	}, {
		name: "initialized",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:   RedisStreamConditionReady,
			Status: corev1.ConditionUnknown,
		},
	}, {
		name: "mark sink",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.MarkSink(apis.HTTP("example").String())
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:   RedisStreamConditionReady,
			Status: corev1.ConditionUnknown,
		},
	}, {
		name: "mark sink and deployed",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.MarkSink(apis.HTTP("example").String())
			s.PropagateStatefulSetAvailability(availableStatefulSet)
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:   RedisStreamConditionReady,
			Status: corev1.ConditionTrue,
		},
	}, {
		name: "mark sink, rolebinding, then no sink",
		s: func() *RedisStreamSourceStatus {
			s := &RedisStreamSourceStatus{}
			s.InitializeConditions()
			s.MarkSink(apis.HTTP("example").String())
			s.MarkRoleBinding()
			s.MarkNoSink("Testing", "hi%s", "")
			return s
		}(),
		condQuery: RedisStreamConditionReady,
		want: &apis.Condition{
			Type:    RedisStreamConditionReady,
			Status:  corev1.ConditionFalse,
			Reason:  "Testing",
			Message: "hi",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.s.GetCondition(test.condQuery)
			ignoreTime := cmpopts.IgnoreFields(apis.Condition{},
				"LastTransitionTime", "Severity")
			if diff := cmp.Diff(test.want, got, ignoreTime); diff != "" {
				t.Errorf("unexpected condition (-want, +got) = %v", diff)
			}
		})
	}
}
