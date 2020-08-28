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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"

	apisv1alpha1 "knative.dev/eventing-redis/pkg/apis/v1alpha1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

// RedisStreamSink is the Schema for the RedisStream API.
type RedisStreamSink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisStreamSinkSpec   `json:"spec,omitempty"`
	Status RedisStreamSinkStatus `json:"status,omitempty"`
}

// Check the interfaces that PingSource should be implementing.
var (
	_ runtime.Object     = (*RedisStreamSink)(nil)
	_ kmeta.OwnerRefable = (*RedisStreamSink)(nil)
	//_ apis.Validatable   = (*RedisStreamSink)(nil)
	//_ apis.Defaultable   = (*RedisStreamSink)(nil)
	_ apis.HasSpec    = (*RedisStreamSink)(nil)
	_ duckv1.KRShaped = (*RedisStreamSink)(nil)
)

// RedisStreamSinkSpec defines the desired state of the RedisStreamSink.
type RedisStreamSinkSpec struct {
	// RedisConnection represents the address and options to connect
	// to a Redis instance
	apisv1alpha1.RedisConnection `json:",inline"`

	// Stream is the name of the stream to send events to
	Stream string `json:"stream"`
}

// RedisStreamSinkStatus defines the observed state of RedisStreamSink.
type RedisStreamSinkStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`
	// AddressStatus is the part where this Sink fulfills the Addressable contract.
	duckv1.AddressStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisStreamSinkList contains a list of RedisStreamSink.
type RedisStreamSinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisStreamSink `json:"items"`
}

// GetStatus retrieves the status of the RedisStreamSink. Implements the KRShaped interface.
func (p *RedisStreamSink) GetStatus() *duckv1.Status {
	return &p.Status.Status
}
