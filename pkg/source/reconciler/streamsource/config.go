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
package streamsource

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
)

const (
	configMapNameEnv    = "CONFIG_REDIS_NUMCONSUMERS"
	redisConfigKey      = "numConsumers"
	DefaultNumConsumers = "10"
	tlsSecretNameEnv    = "SECRET_TLS_TLSCERTIFICATE"
	tlsConfigKey        = "TLS_CERT"
)

// RedisConfig contains the configuration defined in the redis ConfigMap.
// +k8s:deepcopy-gen=true
type RedisConfig struct {
	NumConsumers string
}

type TLSConfig struct {
	TLSCertificate string
}

func defaultConfig() *RedisConfig {
	return &RedisConfig{
		NumConsumers: DefaultNumConsumers,
	}
}

// ConfigMapName gets the name of the redis ConfigMap
func ConfigMapName() string {
	cm := os.Getenv(configMapNameEnv)
	if cm == "" {
		return "config-redis"
	}
	return cm
}

// TLSSecretName gets the name of the tls cert Secret
func TLSSecretName() string {
	cm := os.Getenv(tlsSecretNameEnv)
	if cm == "" {
		return "tls-secret"
	}
	return cm
}

// NewConfigFromMap creates a RedisConfig from the supplied map,
// expecting the given list of components.
func NewConfigFromMap(data map[string]string) (*RedisConfig, error) {
	rc := defaultConfig()
	if numC, ok := data[redisConfigKey]; ok {
		rc.NumConsumers = numC
	}
	return rc, nil
}

// NewConfigFromConfigMap creates a RedisConfig from the supplied ConfigMap,
// expecting the given list of components.
func NewConfigFromConfigMap(configMap *corev1.ConfigMap) (*RedisConfig, error) {
	return NewConfigFromMap(configMap.Data)
}

// GetRedisConfig returns the details of the Redis stream.
func GetRedisConfig(configMap map[string]string) (*RedisConfig, error) {
	if len(configMap) == 0 {
		return nil, fmt.Errorf("missing configuration")
	}

	config := &RedisConfig{
		NumConsumers: DefaultNumConsumers,
	}

	err := configmap.Parse(configMap,
		configmap.AsString(redisConfigKey, &config.NumConsumers),
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetTLSSecret returns the details of the TLS certificate.
func GetTLSSecret(secret map[string][]byte) (*TLSConfig, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("missing configuration")
	}

	//Convert byte to a string
	config := &TLSConfig{
		TLSCertificate: string(secret[tlsConfigKey]),
	}

	if config.TLSCertificate == "" {
		return nil, fmt.Errorf("tls certificate missing from secret")
	}

	return config, nil
}
