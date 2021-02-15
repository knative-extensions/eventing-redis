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
package streamsink

import (
	"fmt"
	"os"

	"knative.dev/pkg/configmap"
)

const (
	tlsConfigMapNameEnv = "CONFIG_TLS_TLSCERTIFICATE"
	tlsConfigKey        = "cert.pem"
)

type TLSConfig struct {
	TLSCertificate string
}

// TLSConfigMapName gets the name of the tls cert ConfigMap
func TLSConfigMapName() string {
	cm := os.Getenv(tlsConfigMapNameEnv)
	if cm == "" {
		return "config-tls"
	}
	return cm
}

// GetTLSConfig returns the details of the TLS certificate.
func GetTLSConfig(configMap map[string]string) (*TLSConfig, error) {
	if len(configMap) == 0 {
		return nil, fmt.Errorf("missing configuration")
	}

	config := &TLSConfig{
		TLSCertificate: "",
	}

	err := configmap.Parse(configMap,
		configmap.AsString(tlsConfigKey, &config.TLSCertificate),
	)
	if err != nil {
		return nil, err
	}

	return config, nil
}
