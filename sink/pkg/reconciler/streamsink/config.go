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
)

const (
	tlsSecretNameEnv = "SECRET_TLS_TLSCERTIFICATE"
	tlsConfigKey     = "TLS_CERT"
)

type TLSConfig struct {
	TLSCertificate string
}

// TLSSecretName gets the name of the tls cert Secret
func TLSSecretName() string {
	cm := os.Getenv(tlsSecretNameEnv)
	if cm == "" {
		return "tls-secret"
	}
	return cm
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
