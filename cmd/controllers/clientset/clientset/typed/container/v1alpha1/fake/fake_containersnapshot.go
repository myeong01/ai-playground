/*
Copyright 2023.

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

	v1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/container/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeContainerSnapshots implements ContainerSnapshotInterface
type FakeContainerSnapshots struct {
	Fake *FakeContainerV1alpha1
	ns   string
}

var containersnapshotsResource = v1alpha1.SchemeGroupVersion.WithResource("containersnapshots")

var containersnapshotsKind = v1alpha1.SchemeGroupVersion.WithKind("ContainerSnapshot")

// Get takes name of the containerSnapshot, and returns the corresponding containerSnapshot object, and an error if there is any.
func (c *FakeContainerSnapshots) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ContainerSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(containersnapshotsResource, c.ns, name), &v1alpha1.ContainerSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ContainerSnapshot), err
}

// List takes label and field selectors, and returns the list of ContainerSnapshots that match those selectors.
func (c *FakeContainerSnapshots) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ContainerSnapshotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(containersnapshotsResource, containersnapshotsKind, c.ns, opts), &v1alpha1.ContainerSnapshotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ContainerSnapshotList{ListMeta: obj.(*v1alpha1.ContainerSnapshotList).ListMeta}
	for _, item := range obj.(*v1alpha1.ContainerSnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested containerSnapshots.
func (c *FakeContainerSnapshots) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(containersnapshotsResource, c.ns, opts))

}

// Create takes the representation of a containerSnapshot and creates it.  Returns the server's representation of the containerSnapshot, and an error, if there is any.
func (c *FakeContainerSnapshots) Create(ctx context.Context, containerSnapshot *v1alpha1.ContainerSnapshot, opts v1.CreateOptions) (result *v1alpha1.ContainerSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(containersnapshotsResource, c.ns, containerSnapshot), &v1alpha1.ContainerSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ContainerSnapshot), err
}

// Update takes the representation of a containerSnapshot and updates it. Returns the server's representation of the containerSnapshot, and an error, if there is any.
func (c *FakeContainerSnapshots) Update(ctx context.Context, containerSnapshot *v1alpha1.ContainerSnapshot, opts v1.UpdateOptions) (result *v1alpha1.ContainerSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(containersnapshotsResource, c.ns, containerSnapshot), &v1alpha1.ContainerSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ContainerSnapshot), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeContainerSnapshots) UpdateStatus(ctx context.Context, containerSnapshot *v1alpha1.ContainerSnapshot, opts v1.UpdateOptions) (*v1alpha1.ContainerSnapshot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(containersnapshotsResource, "status", c.ns, containerSnapshot), &v1alpha1.ContainerSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ContainerSnapshot), err
}

// Delete takes name of the containerSnapshot and deletes it. Returns an error if one occurs.
func (c *FakeContainerSnapshots) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(containersnapshotsResource, c.ns, name, opts), &v1alpha1.ContainerSnapshot{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeContainerSnapshots) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(containersnapshotsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ContainerSnapshotList{})
	return err
}

// Patch applies the patch and returns the patched containerSnapshot.
func (c *FakeContainerSnapshots) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ContainerSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(containersnapshotsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ContainerSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ContainerSnapshot), err
}
