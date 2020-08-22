/*
Copyright 2019 The Knative Authors

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

package adapter

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

const (
	// RedisStreamSourceEventType is the default RedisStreamSource CloudEvent type.
	RedisStreamSourceEventType = "dev.knative.sources.redisstream"
)

func NewEnvConfig() adapter.EnvConfigAccessor {
	return &Config{}
}

type Adapter struct {
	config *Config
	logger *zap.Logger
	client cloudevents.Client
	source string
}

func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	config := processed.(*Config)
	return &Adapter{
		config: config,
		logger: logging.FromContext(ctx).Desugar(),
		client: ceClient,
		source: fmt.Sprintf("%s/%s", config.Address, config.Stream),
	}
}

func (a *Adapter) Start(ctx context.Context) error {
	conn, err := redis.Dial("tcp", a.config.Address)
	if err != nil {
		return err
	}

	a.logger.Info("Listening stream", zap.String("name", a.config.Stream))

	for {
		reply, err := redis.Values(conn.Do("XREAD", "COUNT", 1, "BLOCK", 0, "STREAMS", a.config.Stream, "$"))

		if err != nil {
			a.logger.Error("cannot read from stream", zap.Error(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		event, err := a.toEvent(reply)
		if err != nil {
			a.logger.Error("cannot convert reply", zap.Error(err))
			continue
		}

		if result := a.client.Send(ctx, *event); !cloudevents.IsACK(result) {
			//  Event is lost.
			a.logger.Error("failed to send cloudevent", zap.Any("result", result))
		}
	}
}

func (a *Adapter) toEvent(values []interface{}) (*cloudevents.Event, error) {
	// Assert only one stream
	if len(values) != 1 {
		return nil, fmt.Errorf("number of values not equal to one (got %d)", len(values))
	}
	streamValues := values[0]

	a.logger.Info("streamValues", zap.Any("streamValues", streamValues))

	streamValue, err := redis.Values(streamValues, nil)
	if err != nil {
		return nil, err
	}
	a.logger.Info("streamValue", zap.Any("streamValue", streamValue))

	stream, err := redis.String(streamValue[0], nil)
	if err != nil {
		return nil, err
	}
	a.logger.Info("stream", zap.Any("stream", stream))

	elems, err := redis.Values(streamValue[1], nil)
	if err != nil {
		return nil, err
	}

	// Assert only one element
	if len(elems) != 1 {
		return nil, fmt.Errorf("number of elementss not equal to one (got %d)", len(elems))
	}

	idelem, err := redis.Values(elems[0], nil)
	if err != nil {
		return nil, err
	}

	id, err := redis.String(idelem[0], nil)
	if err != nil {
		return nil, err
	}

	a.logger.Info("id", zap.Any("id", id))

	fieldvalues, err := redis.Values(idelem[1], nil)
	if err != nil {
		return nil, err
	}

	for _, fieldvalue := range fieldvalues {
		v, err := redis.String(fieldvalue, nil)
		if err != nil {
			return nil, err
		}
		a.logger.Info("field", zap.Any("field", v))
	}

	event := cloudevents.NewEvent()
	event.SetType(RedisStreamSourceEventType)
	event.SetSource(a.source)
	event.SetData(cloudevents.ApplicationJSON, fieldvalues)
	event.SetID(id)

	return &event, nil
}
