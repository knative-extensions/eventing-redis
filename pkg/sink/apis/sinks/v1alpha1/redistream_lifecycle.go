/*
Copyright 2020 The Knative Authors

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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	// RedisStreamConditionReady has status True when the RedisStreamSink is ready to send events.
	RedisStreamConditionReady = apis.ConditionReady

	// RedisStreamConditionServiceReady has status True when the RedisStreamSink has had it's Knative service created and ready
	RedisStreamConditionServiceReady apis.ConditionType = "ServiceReady"
)

var redisStreamCondSet = apis.NewLivingConditionSet(RedisStreamConditionServiceReady)

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*RedisStreamSink) GetConditionSet() apis.ConditionSet {
	return redisStreamCondSet
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *RedisStreamSink) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("RedisStreamSink")
}

// GetUntypedSpec returns the spec of the RedisStreamSink.
func (s *RedisStreamSink) GetUntypedSpec() interface{} {
	return s.Spec
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *RedisStreamSinkStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return redisStreamCondSet.Manage(s).GetCondition(t)
}

// GetTopLevelCondition returns the top level condition.
func (s *RedisStreamSinkStatus) GetTopLevelCondition() *apis.Condition {
	return redisStreamCondSet.Manage(s).GetTopLevelCondition()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *RedisStreamSinkStatus) InitializeConditions() {
	redisStreamCondSet.Manage(s).InitializeConditions()
}

// IsReady returns true if the resource is ready overall.
func (s *RedisStreamSinkStatus) IsReady() bool {
	return redisStreamCondSet.Manage(s).IsHappy()
}

// PropagateKnativeServiceAddress propagates the Ksvc address to the sink
func (s *RedisStreamSinkStatus) PropagateKnativeServiceAddress(ks *servingv1.Service) bool {
	if ks.Status.GetCondition(apis.ConditionReady).IsTrue() && ks.Status.Address != nil {
		s.Address = ks.Status.Address
		redisStreamCondSet.Manage(s).MarkTrue(RedisStreamConditionServiceReady)
		return true
	}
	return false
}

// MarkNoRoleBinding sets the annotation that the sink does not have a role binding
func (s *RedisStreamSinkStatus) MarkNoRoleBinding(reason string) {
	s.setAnnotation("roleBinding", reason)
}

// MarkRoleBinding sets the annotation that the sink has a role binding
func (s *RedisStreamSinkStatus) MarkRoleBinding() {
	s.clearAnnotation("roleBinding")
}

// MarkNoServiceAccount sets the annotation that the sink does not have a service account
func (s *RedisStreamSinkStatus) MarkNoServiceAccount(reason string) {
	s.setAnnotation("serviceAccount", reason)
}

// MarkRoleBinding sets the annotation that the sink has a service account
func (s *RedisStreamSinkStatus) MarkServiceAccount() {
	s.clearAnnotation("serviceAccount")
}

// MarkNoKnativeService sets the annotation that the sink does not have a Knative Service
func (s *RedisStreamSinkStatus) MarkNoKnativeService(reason string) {
	s.setAnnotation("knativeService", reason)
}

// MarkKnativeService sets the annotation that the sink has a Knative Service
func (s *RedisStreamSinkStatus) MarkKnativeService() {
	s.clearAnnotation("knativeService")
}

func (s *RedisStreamSinkStatus) setAnnotation(name, value string) {
	if s.Annotations == nil {
		s.Annotations = make(map[string]string)
	}
	s.Annotations[name] = value
}

func (s *RedisStreamSinkStatus) clearAnnotation(name string) {
	delete(s.Annotations, name)
}
