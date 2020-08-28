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

import corev1 "k8s.io/api/core/v1"

// RedisConnection defines the address and options to connect to a Redis instance
// +k8s:deepcopy-gen=true
type RedisConnection struct {
	// Address is the Redis TCP address
	Address string `json:"address"`

	// Options are the connection options
	// +optional
	Options *RedisConnectionOptions `json:"dialOptions,omitempty"`
}

// RedisConnection defines the desired state of the RedisStreamSource.
// +k8s:deepcopy-gen=true
type RedisConnectionOptions struct {
	// Password to use for connecting to Redis
	// +optional
	Password corev1.ObjectReference `json:"password,omitempty"`

	// UseTLS indicates whether to use TLS or not
	// +optional
	UseTLS bool `json:"useTLS,omitempty"`

	// SkipVerify indicates whether to skip TLS verification or not
	// +optional
	SkipVerify bool `json:"skipVerify,omitempty"`

	// Cert is the Kubernetes secret containing the client certificate.
	// +optional
	Cert RedisSecretValueFromSource `json:"cert,omitempty"`

	// Key is the Kubernetes secret containing the client key.
	// +optional
	Key RedisSecretValueFromSource `json:"key,omitempty"`

	// CACert is the Kubernetes secret containing the server CA cert.
	// +optional
	CACert RedisSecretValueFromSource `json:"caCert,omitempty"`
}

// RedisSecretValueFromSource represents the source of a secret value
// +k8s:deepcopy-gen=true
type RedisSecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
