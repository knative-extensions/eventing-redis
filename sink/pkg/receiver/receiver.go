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

package receiver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	redisParse "github.com/go-redis/redis/v8"
)

type Receiver interface {
	Receive(event cloudevents.Event)
}

type receiver struct {
	config *Config
	logger *zap.Logger
	pool   *redis.Pool
}

func NewEnvConfig() adapter.EnvConfigAccessor {
	return &Config{}
}

func NewReceiver(ctx context.Context, processed adapter.EnvConfigAccessor) Receiver {
	config := processed.(*Config)
	return &receiver{
		config: config,
		pool:   newPool(config.Address, config.TLSCertificate),
		logger: logging.FromContext(ctx).Desugar().With(zap.String("stream", config.Stream)),
	}
}

func (r *receiver) Receive(event cloudevents.Event) {
	r.logger.Info("Receiving event", zap.Any("event", event))
	conn, _ := r.pool.Dial()
	defer conn.Close()

	// TODO: validate event
	var fields []interface{}
	err := json.Unmarshal(event.Data(), &fields)
	if err != nil {
		r.logger.Error("Cannot decode event", zap.Error(err))
		return
	}

	args := []interface{}{r.config.Stream, "*"}
	args = append(args, fields...)

	_, err = conn.Do("XADD", args...)
	if err != nil {
		r.logger.Error("Cannot write to stream", zap.Error(err))
	}
	r.logger.Info("Added event to the stream")

}

func newPool(address string, tlscert string) *redis.Pool {
	opt, err := redisParse.ParseURL(address)
	if err != nil {
		panic(err)
	}
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			var c redis.Conn
			if opt.Password != "" && tlscert != "" {
				roots := x509.NewCertPool()
				ok := roots.AppendCertsFromPEM([]byte(tlscert))
				if !ok {
					panic(err)
				}
				c, err = redis.Dial("tcp", opt.Addr,
					//redis.DialUsername(opt.Username), //username needs to be empty for successful redis connection (v8 go-redis issue)
					redis.DialPassword(opt.Password),
					redis.DialTLSConfig(&tls.Config{
						RootCAs: roots,
					}),
					redis.DialTLSSkipVerify(true),
					redis.DialUseTLS(true),
				)
				if err != nil {
					panic(err)
				}
			} else {
				c, err = redis.Dial("tcp", opt.Addr)
				if err != nil {
					panic(err)
				}
			}
			return c, err
		},
	}
}
