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
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
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
		pool:   newPool(config.Address),
		logger: logging.FromContext(ctx).Desugar().With(zap.String("stream", config.Stream)),
	}
}

func (r *receiver) Receive(event cloudevents.Event) {
	r.logger.Info("receiving event", zap.Any("event", event))
	conn, _ := r.pool.Dial()
	defer conn.Close()

	// TODO: validate event
	var fields []interface{}
	err := json.Unmarshal(event.Data(), &fields)
	if err != nil {
		r.logger.Error("cannot decode event", zap.Error(err))
		return
	}
	args := []interface{}{r.config.Stream, "*"}
	args = append(args, fields...)

	_, err = conn.Do("XADD", args...)

	if err != nil {
		r.logger.Error("cannot write to stream", zap.Error(err))
	}
}

func newPool(address string) *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}
