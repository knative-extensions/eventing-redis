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
	pool := newPool(a.config.Address)

	conn, err := pool.Dial()
	if err != nil {
		return err
	}

	groupName := a.config.Group             //TODO: If empty, need to use a default/random group name?
	maxPenCount := a.config.MaxPendingCount //TODO: If empty, need to use a default count?

	a.logger.Info("Retrieving group info", zap.String("group", groupName))
	groups, err := scan.ScanXInfoGroupReply(conn.Do("XINFO", "GROUPS", a.config.Stream))
	if err != nil {
		return err
	}

	if groupInfo, ok := groups[groupName]; ok {
		a.logger.Info("Reusing consumer group", zap.String("group", groupName))

		if groupInfo.Pending > 0 {
			// Process pending messages that may be permanently failing
			pendingmsgs, err := scan.ScanXPendingReply(conn.Do("XPENDING", a.config.Stream, groupName, "-", "+", maxPenCount))
			if err != nil {
				return err
			}

			if len(pendingmsgs) > 0 {
				// TODO: check idletime for each message and xclaim to different consumer than current owner? THINK ABOUT DESIGN
			}

		}

	} else {
		a.logger.Info("Creating consumer group", zap.String("group", groupName))
		_, err := conn.Do("XGROUP", "CREATE", a.config.Stream, groupName, "$")
		if err != nil {
			return err
		}
	}
	conn.Close()

	for i := 0; i < 100; i++ {
		go func() {
			conn, _ := pool.Dial()

			// TODO: get it from statefulset pod name
			consumerName := fmt.Sprintf("consumer-%d", i)

			a.logger.Info("Listening for messages", zap.String("consumerName", consumerName))

			for {
				//following xread only reads new data (so what about when consumer fails?)
				reply, err := conn.Do("XREADGROUP", "GROUP", groupName, consumerName, "COUNT", 1, "BLOCK", 0, "STREAMS", a.config.Stream, ">")

				if err != nil {
					a.logger.Error("cannot read from stream", zap.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				a.logger.Info("consumer reading from stream", zap.String("consumerName", consumerName))

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
				a.logger.Info("consumer acknowledged the message", zap.String("consumerName", consumerName))
			}
		}()
	}

	<-ctx.Done()
	return nil
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
