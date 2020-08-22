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

package scan

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

//1) 1) "mystream"
//   2) 1) 1) 1519073278252-0
//         2) 1) "foo"
//            2) "value_1"
//      2) 1) 1519073279157-0
//         2) 1) "foo"
//            2) "value_2"

type StreamElements []StreamElement

type StreamElement struct {
	// Name is the stream name
	Name string

	// Items is the stream items (ID and list of field-value pairs)
	Items []StreamItem
}

type StreamItem struct {
	// ID is the item ID
	ID string

	// FieldValue represent the unscan list of field-value pairs
	FieldValues interface{}
}

func ScanStreamResult(src []interface{}, dst StreamElements) (StreamElements, error) {
	if dst == nil || len(dst) != len(src) {
		se := make(StreamElements, len(src))
		dst = se
	}

	for i, stream := range src {
		//		a.logger.Info("streamValues", zap.Any("streamValues", streamValues))
		elem, err := redis.Values(stream, nil)
		if err != nil {
			return nil, err
		}
		if len(elem) != 2 {
			return nil, fmt.Errorf("unexpected stream element slice length (%d)", len(elem))
		}

		// a.logger.Info("streamValue", zap.Any("streamValue", streamValue))

		name, err := redis.String(elem[0], nil)
		if err != nil {
			return nil, err
		}
		//a.logger.Info("stream", zap.Any("stream", stream))

		dst[i].Name = name

		items, err := redis.Values(elem[1], nil)
		if err != nil {
			return nil, err
		}

		if len(dst[i].Items) != len(items) {
			// Reallocate
			dst[i].Items = make([]StreamItem, len(items))
		}

		for j, rawitem := range items {
			item, err := redis.Values(rawitem, nil)
			if err != nil {
				return nil, err
			}

			if len(item) != 2 {
				return nil, fmt.Errorf("unexpected stream item slice length (%d)", len(elem))
			}

			id, err := redis.String(item[0], nil)
			if err != nil {
				return nil, err
			}
			dst[i].Items[j].ID = id
			dst[i].Items[j].FieldValues = item[1]
		}
	}
	return dst, nil
}
