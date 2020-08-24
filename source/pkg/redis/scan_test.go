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
	"testing"

	"github.com/google/go-cmp/cmp"
)

//1) 1) "mystream"
//   2) 1) 1) 1519073278252-0
//         2) 1) "foo"
//            2) "value_1"
//      2) 1) 1519073279157-0
//         2) 1) "foo"
//            2) "value_2"

func TestScanXRead(t *testing.T) {
	tests := []struct {
		reply    []interface{}
		expected StreamElements
	}{
		{
			reply: []interface{}{
				[]interface{}{
					[]byte("mystream"),
					[]interface{}{
						[]interface{}{
							[]byte("1519073278252-0"),
							[]interface{}{
								[]byte("foo"),
								[]byte("value_1")}}}}},
			expected: []StreamElement{
				{
					Name: "mystream",
					Items: []StreamItem{
						{
							ID: "1519073278252-0",
							FieldValues: []interface{}{
								[]byte("foo"),
								[]byte("value_1")}},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		actual, err := ScanXReadReply(tc.reply, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if diff := cmp.Diff(tc.expected, actual); diff != "" {
			t.Errorf("Unexpected difference (-want, +got): %v", diff)
		}

	}

}
