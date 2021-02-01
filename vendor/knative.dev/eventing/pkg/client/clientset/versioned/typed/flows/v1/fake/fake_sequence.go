/*
Copyright 2021 The Knative Authors

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	flowsv1 "knative.dev/eventing/pkg/apis/flows/v1"
)

// FakeSequences implements SequenceInterface
type FakeSequences struct {
	Fake *FakeFlowsV1
	ns   string
}

var sequencesResource = schema.GroupVersionResource{Group: "flows.knative.dev", Version: "v1", Resource: "sequences"}

var sequencesKind = schema.GroupVersionKind{Group: "flows.knative.dev", Version: "v1", Kind: "Sequence"}

// Get takes name of the sequence, and returns the corresponding sequence object, and an error if there is any.
func (c *FakeSequences) Get(ctx context.Context, name string, options v1.GetOptions) (result *flowsv1.Sequence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sequencesResource, c.ns, name), &flowsv1.Sequence{})

	if obj == nil {
		return nil, err
	}
	return obj.(*flowsv1.Sequence), err
}

// List takes label and field selectors, and returns the list of Sequences that match those selectors.
func (c *FakeSequences) List(ctx context.Context, opts v1.ListOptions) (result *flowsv1.SequenceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sequencesResource, sequencesKind, c.ns, opts), &flowsv1.SequenceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &flowsv1.SequenceList{ListMeta: obj.(*flowsv1.SequenceList).ListMeta}
	for _, item := range obj.(*flowsv1.SequenceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sequences.
func (c *FakeSequences) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sequencesResource, c.ns, opts))

}

// Create takes the representation of a sequence and creates it.  Returns the server's representation of the sequence, and an error, if there is any.
func (c *FakeSequences) Create(ctx context.Context, sequence *flowsv1.Sequence, opts v1.CreateOptions) (result *flowsv1.Sequence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sequencesResource, c.ns, sequence), &flowsv1.Sequence{})

	if obj == nil {
		return nil, err
	}
	return obj.(*flowsv1.Sequence), err
}

// Update takes the representation of a sequence and updates it. Returns the server's representation of the sequence, and an error, if there is any.
func (c *FakeSequences) Update(ctx context.Context, sequence *flowsv1.Sequence, opts v1.UpdateOptions) (result *flowsv1.Sequence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sequencesResource, c.ns, sequence), &flowsv1.Sequence{})

	if obj == nil {
		return nil, err
	}
	return obj.(*flowsv1.Sequence), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSequences) UpdateStatus(ctx context.Context, sequence *flowsv1.Sequence, opts v1.UpdateOptions) (*flowsv1.Sequence, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(sequencesResource, "status", c.ns, sequence), &flowsv1.Sequence{})

	if obj == nil {
		return nil, err
	}
	return obj.(*flowsv1.Sequence), err
}

// Delete takes name of the sequence and deletes it. Returns an error if one occurs.
func (c *FakeSequences) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sequencesResource, c.ns, name), &flowsv1.Sequence{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSequences) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sequencesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &flowsv1.SequenceList{})
	return err
}

// Patch applies the patch and returns the patched sequence.
func (c *FakeSequences) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *flowsv1.Sequence, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sequencesResource, c.ns, name, pt, data, subresources...), &flowsv1.Sequence{})

	if obj == nil {
		return nil, err
	}
	return obj.(*flowsv1.Sequence), err
}
