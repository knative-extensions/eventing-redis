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
	"strings"
	"sync"
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
	blockms                    = 5000 // block for 5seconds before timing out
	count                      = 1    //read one redis entry at a time
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

	waitGroup := &sync.WaitGroup{}
	pool := newPool(a.config.Address)

	conn, err := pool.Dial()
	if err != nil {
		return err
	}

	streamName := a.config.Stream
	groupName := a.config.PodName // Build consumer group name from stateful set pod name of adapter

	a.logger.Info("Retrieving group info", zap.String("group", groupName))
	groups, err := scan.ScanXInfoGroupReply(conn.Do("XINFO", "GROUPS", streamName))

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such key") || strings.Contains(strings.ToLower(err.Error()), "no longer exists") {
			// stream does not exist, may have been deleted accidentally
			a.logger.Info("Creating stream and consumer group", zap.String("group", groupName))
			//XGROUP CREATE creates the stream automatically, if it doesn't exist, when MKSTREAM subcommand is specified as last argument
			_, err := conn.Do("XGROUP", "CREATE", streamName, groupName, "$", "MKSTREAM")
			if err != nil {
				a.logger.Error("Cannot create stream and consumer group", zap.Error(err))
				return err
			}
		} else {
			return err
		}

	} else {

		if _, ok := groups[groupName]; ok {
			a.logger.Info("Reusing consumer group", zap.String("group", groupName))
		} else {
			a.logger.Info("Creating consumer group", zap.String("group", groupName))
			_, err := conn.Do("XGROUP", "CREATE", streamName, groupName, "$")
			if err != nil {
				a.logger.Error("Cannot create consumer group", zap.Error(err))
				return err
			}
		}

	}

	a.logger.Info("Number of consumers from config:", zap.Int("NumConsumers", a.config.NumConsumers))

	for i := 0; i < a.config.NumConsumers; i++ {
		waitGroup.Add(1)

		go func(wg *sync.WaitGroup, j int) {
			defer wg.Done()

			conn, _ := pool.Dial()

			consumerName := fmt.Sprintf("%s-%d", groupName, j)
			xreadID := "0" //Initial ID to read pending messages
			a.logger.Info("Listening for messages", zap.String("consumerName", consumerName))

			for {
				select {
				case <-ctx.Done(): //received a SIGINT or SIGTERM signal. Need to process pending messages and shut down consumer group

					morePendingMsgs := true
					for morePendingMsgs {
						reply, err := conn.Do("XREADGROUP", "GROUP", groupName, consumerName, "COUNT", count, "BLOCK", blockms, "STREAMS", streamName, xreadID)
						if err != nil {
							a.logger.Error("Cannot read from stream", zap.Error(err))
							continue
						}

						event, err := a.toEvent(reply)
						if err != nil {
							if strings.Contains(strings.ToLower(err.Error()), "number of items not equal to one (got 0)") || // no more pending messages!
								strings.Contains(strings.ToLower(err.Error()), "expected a reply of type array") { // Xreadgroup timed out blocking after 5 seconds
								morePendingMsgs = false
							} else {
								a.logger.Error("Cannot convert reply", zap.Error(err))
							}
							continue
						}

						if result := a.client.Send(ctx, *event); !cloudevents.IsACK(result) { //  Event is lost
							a.logger.Error("Failed to send cloudevent", zap.Any("result", result))
						}

						_, err = conn.Do("XACK", streamName, groupName, event.ID())
						if err != nil {
							a.logger.Error("Cannot ack message", zap.Error(err))
							morePendingMsgs = true
							continue
						}
						a.logger.Info("Consumer acknowledged the message", zap.String("consumerName", consumerName))
					}

					_, err := conn.Do("XGROUP", "DELCONSUMER", streamName, groupName, consumerName)
					if err != nil {
						a.logger.Error("Cannot delete consumer", zap.Error(err))
					}

					a.logger.Info("Consumer shut down", zap.String("consumerName", consumerName))

					conn.Close()
					return
				default:
					//XREAD reads all the pending messages when xreadID=="0" and new messages when xreadID==">"
					reply, err := conn.Do("XREADGROUP", "GROUP", groupName, consumerName, "COUNT", count, "BLOCK", blockms, "STREAMS", streamName, xreadID)
					if err != nil {
						a.logger.Error("Cannot read from stream", zap.Error(err))
						time.Sleep(1 * time.Second)
						continue
					}

					event, err := a.toEvent(reply)
					if err != nil {
						if strings.Contains(strings.ToLower(err.Error()), "number of items not equal to one (got 0)") || // no more pending messages!
							strings.Contains(strings.ToLower(err.Error()), "expected a reply of type array") { // Xreadgroup timed out blocking after 5 seconds
							xreadID = ">" //ID to read new messages in next iteration
						} else {
							a.logger.Error("Cannot convert reply", zap.Error(err))
							time.Sleep(1 * time.Second)
						}
						continue
					}

					a.logger.Info("Consumer read a message", zap.String("consumerName", consumerName))

					if result := a.client.Send(ctx, *event); !cloudevents.IsACK(result) { //  Event is lost
						a.logger.Error("Failed to send cloudevent", zap.Any("result", result))
					}

					_, err = conn.Do("XACK", streamName, groupName, event.ID())
					if err != nil {
						a.logger.Error("Cannot ack message", zap.Error(err))
						xreadID = "0" //ID to read pending message in next iteration
						time.Sleep(1 * time.Second)
						continue
					}
					a.logger.Info("Consumer acknowledged the message", zap.String("consumerName", consumerName))
				}
			}
		}(waitGroup, i)
	}

	waitGroup.Wait() // wait for all consumers

	a.logger.Info("Quit signal received, gracefully shutdown all consumers.")

	_, err = conn.Do("XGROUP", "DESTROY", streamName, groupName)
	if err != nil {
		a.logger.Error("Cannot destroy consumer group", zap.Error(err))
		return err
	}
	conn.Close()

	a.logger.Info("Done. All consumers are stopped now.")

	return nil
}

func (a *Adapter) processEntry(ctx context.Context, conn redis.Conn, streamName string, groupName string, consumerName string, xreadID int) {
	reply, err := conn.Do("XREADGROUP", "GROUP", groupName, consumerName, "COUNT", count, "BLOCK", blockms, "STREAMS", streamName, xreadID)
	if err != nil {
		a.logger.Error("Cannot read from stream", zap.Error(err))
		time.Sleep(1 * time.Second)
		continue
	}

	event, err := a.toEvent(reply)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "number of items not equal to one (got 0)") || // no more pending messages!
			strings.Contains(strings.ToLower(err.Error()), "expected a reply of type array") { // Xreadgroup timed out blocking after 5 seconds
			xreadID = ">" //ID to read new messages in next iteration
		} else {
			a.logger.Error("Cannot convert reply", zap.Error(err))
			time.Sleep(1 * time.Second)
		}
		continue
	}

	a.logger.Info("Consumer read a message", zap.String("consumerName", consumerName))

	if result := a.client.Send(ctx, *event); !cloudevents.IsACK(result) { //  Event is lost
		a.logger.Error("Failed to send cloudevent", zap.Any("result", result))
	}

	_, err = conn.Do("XACK", streamName, groupName, event.ID())
	if err != nil {
		a.logger.Error("Cannot ack message", zap.Error(err))
		xreadID = "0" //ID to read pending message in next iteration
		time.Sleep(1 * time.Second)
		continue
	}
	a.logger.Info("Consumer acknowledged the message", zap.String("consumerName", consumerName))
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
