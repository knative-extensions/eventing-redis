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

package adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	scan "knative.dev/eventing-redis/source/pkg/redis"

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
		logger: logging.FromContext(ctx).Desugar().With(zap.String("stream", config.Stream)),
		client: ceClient,
		source: fmt.Sprintf("%s/%s", config.Address, config.Stream),
	}
}

func (a *Adapter) Start(ctx context.Context) error {
	conn, err := redis.Dial("tcp", a.config.Address)
	if err != nil {
		return err
	}

	// TODO: get it from spec
	groupName := "group"

	a.logger.Info("Retrieving group info", zap.String("group", groupName))
	groups, err := scan.ScanXInfoGroupReply(conn.Do("XINFO", "GROUPS", a.config.Stream))
	if err != nil {
		return err
	}

	if _, ok := groups[groupName]; ok {
		a.logger.Info("Reusing consumer group", zap.String("group", groupName))
		// TODO: process pending messages
	} else {
		a.logger.Info("Creating consumer group", zap.String("group", groupName))
		_, err := conn.Do("XGROUP", "CREATE", a.config.Stream, groupName, "$")
		if err != nil {
			return err
		}
	}

	// TODO: get it from statefulset pod name
	consumerName := "consumer-0"

	a.logger.Info("Listening for messages")

	for {
		reply, err := conn.Do("XREADGROUP", "GROUP", groupName, consumerName, "COUNT", 1, "BLOCK", 0, "STREAMS", a.config.Stream, ">")

		if err != nil {
			a.logger.Error("cannot read from stream", zap.Error(err))
			time.Sleep(1 * time.Second)
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

		_, err = conn.Do("XACK", a.config.Stream, groupName, event.ID())
		if err != nil {
			a.logger.Error("cannot ack message", zap.Error(err))
		}
	}
}

func (a *Adapter) toEvent(reply interface{}) (*cloudevents.Event, error) {
	values, err := redis.Values(reply, nil)
	if err != nil {
		return nil, errors.New("expected a reply of type array")
	}

	// Assert only one stream
	if len(values) != 1 {
		return nil, fmt.Errorf("number of values not equal to one (got %d)", len(values))
	}

	elems, err := scan.ScanXReadReply(values, nil)
	if err != nil {
		return nil, err
	}

	// Assert only one item
	if len(elems[0].Items) != 1 {
		return nil, fmt.Errorf("number of items not equal to one (got %d)", len(elems[0].Items))
	}

	item := elems[0].Items[0]

	event := cloudevents.NewEvent()
	event.SetType(RedisStreamSourceEventType)
	event.SetSource(a.source)
	event.SetData(cloudevents.ApplicationJSON, item.FieldValues)
	event.SetID(item.ID)

	return &event, nil
}
