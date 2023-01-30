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

	v1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/image/v1alpha1"
	scheme "github.com/myeong01/ai-playground/cmd/controllers/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ImagesGetter has a method to return a ImageInterface.
// A group's client should implement this interface.
type ImagesGetter interface {
	Images(namespace string) ImageInterface
}

// ImageInterface has methods to work with Image resources.
type ImageInterface interface {
	Create(ctx context.Context, image *v1alpha1.Image, opts v1.CreateOptions) (*v1alpha1.Image, error)
	Update(ctx context.Context, image *v1alpha1.Image, opts v1.UpdateOptions) (*v1alpha1.Image, error)
	UpdateStatus(ctx context.Context, image *v1alpha1.Image, opts v1.UpdateOptions) (*v1alpha1.Image, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Image, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ImageList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Image, err error)
	ImageExpansion
}

// images implements ImageInterface
type images struct {
	client rest.Interface
	ns     string
}

// newImages returns a Images
func newImages(c *ImageV1alpha1Client, namespace string) *images {
	return &images{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the image, and returns the corresponding image object, and an error if there is any.
func (c *images) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Image, err error) {
	result = &v1alpha1.Image{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("images").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Images that match those selectors.
func (c *images) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ImageList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ImageList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("images").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested images.
func (c *images) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("images").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a image and creates it.  Returns the server's representation of the image, and an error, if there is any.
func (c *images) Create(ctx context.Context, image *v1alpha1.Image, opts v1.CreateOptions) (result *v1alpha1.Image, err error) {
	result = &v1alpha1.Image{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("images").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(image).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a image and updates it. Returns the server's representation of the image, and an error, if there is any.
func (c *images) Update(ctx context.Context, image *v1alpha1.Image, opts v1.UpdateOptions) (result *v1alpha1.Image, err error) {
	result = &v1alpha1.Image{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("images").
		Name(image.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(image).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *images) UpdateStatus(ctx context.Context, image *v1alpha1.Image, opts v1.UpdateOptions) (result *v1alpha1.Image, err error) {
	result = &v1alpha1.Image{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("images").
		Name(image.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(image).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the image and deletes it. Returns an error if one occurs.
func (c *images) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("images").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *images) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("images").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched image.
func (c *images) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Image, err error) {
	result = &v1alpha1.Image{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("images").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
