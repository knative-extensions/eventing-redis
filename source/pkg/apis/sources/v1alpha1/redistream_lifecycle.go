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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

const (
	// RedisStreamConditionReady has status True when the RedisStreamSource is ready to send events.
	RedisStreamConditionReady = apis.ConditionReady

	// RedisStreamConditionSinkProvided has status True when the RedisStreamSource has been configured with a sink target.
	RedisStreamConditionSinkProvided apis.ConditionType = "SinkProvided"

	// RedisStreamConditionDeployed has status True when the RedisStreamSource has had it's deployment created.
	RedisStreamConditionDeployed apis.ConditionType = "Deployed"
)

var redisStreamCondSet = apis.NewLivingConditionSet(
	RedisStreamConditionSinkProvided,
	RedisStreamConditionDeployed,
)

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*RedisStreamSource) GetConditionSet() apis.ConditionSet {
	return redisStreamCondSet
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *RedisStreamSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("RedisStreamSource")
}

// GetUntypedSpec returns the spec of the RedisStreamSource.
func (s *RedisStreamSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *RedisStreamSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return redisStreamCondSet.Manage(s).GetCondition(t)
}

// GetTopLevelCondition returns the top level condition.
func (s *RedisStreamSourceStatus) GetTopLevelCondition() *apis.Condition {
	return redisStreamCondSet.Manage(s).GetTopLevelCondition()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *RedisStreamSourceStatus) InitializeConditions() {
	redisStreamCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *RedisStreamSourceStatus) MarkSink(uri string) {
	s.SinkURI = nil
	if len(uri) > 0 {
		if u, err := apis.ParseURL(uri); err != nil {
			redisStreamCondSet.Manage(s).MarkFalse(RedisStreamConditionSinkProvided, "SinkInvalid", "Failed to parse sink: %v", err)
		} else {
			s.SinkURI = u
			redisStreamCondSet.Manage(s).MarkTrue(RedisStreamConditionSinkProvided)
		}

	} else {
		redisStreamCondSet.Manage(s).MarkFalse(RedisStreamConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *RedisStreamSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	redisStreamCondSet.Manage(s).MarkFalse(RedisStreamConditionSinkProvided, reason, messageFormat, messageA...)
}

// PropagateDeploymentAvailability uses the availability of the provided Deployment to determine if
// RedisStreamConditionDeployed should be marked as true or false.
func (s *RedisStreamSourceStatus) PropagateDeploymentAvailability(d *appsv1.Deployment) {
	deploymentAvailableFound := false
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			deploymentAvailableFound = true
			if cond.Status == corev1.ConditionTrue {
				redisStreamCondSet.Manage(s).MarkTrue(RedisStreamConditionDeployed)
			} else if cond.Status == corev1.ConditionFalse {
				redisStreamCondSet.Manage(s).MarkFalse(RedisStreamConditionDeployed, cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				redisStreamCondSet.Manage(s).MarkUnknown(RedisStreamConditionDeployed, cond.Reason, cond.Message)
			}
		}
	}
	if !deploymentAvailableFound {
		redisStreamCondSet.Manage(s).MarkUnknown(RedisStreamConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

// IsReady returns true if the resource is ready overall.
func (s *RedisStreamSourceStatus) IsReady() bool {
	return redisStreamCondSet.Manage(s).IsHappy()
}
