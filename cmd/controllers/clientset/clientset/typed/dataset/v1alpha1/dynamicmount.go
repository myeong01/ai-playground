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

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/dataset/v1alpha1"
	scheme "github.com/myeong01/ai-playground/cmd/controllers/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DynamicMountsGetter has a method to return a DynamicMountInterface.
// A group's client should implement this interface.
type DynamicMountsGetter interface {
	DynamicMounts(namespace string) DynamicMountInterface
}

// DynamicMountInterface has methods to work with DynamicMount resources.
type DynamicMountInterface interface {
	Create(ctx context.Context, dynamicMount *v1alpha1.DynamicMount, opts v1.CreateOptions) (*v1alpha1.DynamicMount, error)
	Update(ctx context.Context, dynamicMount *v1alpha1.DynamicMount, opts v1.UpdateOptions) (*v1alpha1.DynamicMount, error)
	UpdateStatus(ctx context.Context, dynamicMount *v1alpha1.DynamicMount, opts v1.UpdateOptions) (*v1alpha1.DynamicMount, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.DynamicMount, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.DynamicMountList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DynamicMount, err error)
	DynamicMountExpansion
}

// dynamicMounts implements DynamicMountInterface
type dynamicMounts struct {
	client rest.Interface
	ns     string
}

// newDynamicMounts returns a DynamicMounts
func newDynamicMounts(c *DatasetV1alpha1Client, namespace string) *dynamicMounts {
	return &dynamicMounts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the dynamicMount, and returns the corresponding dynamicMount object, and an error if there is any.
func (c *dynamicMounts) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DynamicMount, err error) {
	result = &v1alpha1.DynamicMount{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dynamicmounts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DynamicMounts that match those selectors.
func (c *dynamicMounts) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DynamicMountList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.DynamicMountList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dynamicmounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested dynamicMounts.
func (c *dynamicMounts) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("dynamicmounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a dynamicMount and creates it.  Returns the server's representation of the dynamicMount, and an error, if there is any.
func (c *dynamicMounts) Create(ctx context.Context, dynamicMount *v1alpha1.DynamicMount, opts v1.CreateOptions) (result *v1alpha1.DynamicMount, err error) {
	result = &v1alpha1.DynamicMount{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("dynamicmounts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(dynamicMount).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a dynamicMount and updates it. Returns the server's representation of the dynamicMount, and an error, if there is any.
func (c *dynamicMounts) Update(ctx context.Context, dynamicMount *v1alpha1.DynamicMount, opts v1.UpdateOptions) (result *v1alpha1.DynamicMount, err error) {
	result = &v1alpha1.DynamicMount{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("dynamicmounts").
		Name(dynamicMount.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(dynamicMount).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *dynamicMounts) UpdateStatus(ctx context.Context, dynamicMount *v1alpha1.DynamicMount, opts v1.UpdateOptions) (result *v1alpha1.DynamicMount, err error) {
	result = &v1alpha1.DynamicMount{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("dynamicmounts").
		Name(dynamicMount.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(dynamicMount).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the dynamicMount and deletes it. Returns an error if one occurs.
func (c *dynamicMounts) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dynamicmounts").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *dynamicMounts) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dynamicmounts").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched dynamicMount.
func (c *dynamicMounts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DynamicMount, err error) {
	result = &v1alpha1.DynamicMount{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("dynamicmounts").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
