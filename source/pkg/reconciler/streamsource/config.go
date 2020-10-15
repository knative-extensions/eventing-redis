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
	"os"

	corev1 "k8s.io/api/core/v1"
)

const (
	configMapNameEnv   = "CONFIG_REDIS_NUMCONSUMERS"
	redisConfigKey     = "numConsumers"
)

// Config contains the configuration defined in the redis ConfigMap.
// +k8s:deepcopy-gen=true
type Config struct {
	NumConsumers    string
}

func defaultConfig() *Config {
	return &Config{
		NumConsumers: "5",
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

// NewConfigFromMap creates a RedisConfig from the supplied map,
// expecting the given list of components.
func NewConfigFromMap(data map[string]string) (*Config, error) {
	rc := defaultConfig()
	if numC, ok := data[redisConfigKey]; ok {
		rc.NumConsumers  = numC
	}
	return rc, nil
}

// NewConfigFromConfigMap creates a Config from the supplied ConfigMap,
// expecting the given list of components.
func NewConfigFromConfigMap(configMap *corev1.ConfigMap) (*Config, error) {
	return NewConfigFromMap(configMap.Data)
}