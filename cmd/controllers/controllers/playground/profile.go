package playground

import (
	"context"
	"github.com/kubeflow/kubeflow/components/profile-controller/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
)

type ProfileInterface interface {
	Create(ctx context.Context, profile *v1beta1.Profile) (*v1beta1.Profile, error)
	Delete(ctx context.Context, name string, opts *metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1beta1.Profile, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1beta1.ProfileList, error)
	Update(ctx context.Context, profile *v1beta1.Profile) (*v1beta1.Profile, error)
}

type ProfileClient struct {
	restClient rest.Interface
}

const GroupName = "kubeflow.org"
const GroupVersion = "v1beta1"

func NewProfileClient(conf *rest.Config) (*ProfileClient, error) {
	profileRestClient, err := getRESTClient(conf, GroupName, GroupVersion)
	if err != nil {
		return nil, err
	}
	return &ProfileClient{
		restClient: profileRestClient,
	}, nil
}

func getRESTClient(restconfig *rest.Config, group string, version string) (*rest.RESTClient, error) {
	restconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: group, Version: version}
	restconfig.APIPath = "/apis"
	restconfig.UserAgent = rest.DefaultKubernetesUserAgent()
	return rest.RESTClientFor(restconfig)
}

const Profiles = "profiles"

func (c *ProfileClient) Create(ctx context.Context, profile *v1beta1.Profile) (*v1beta1.Profile, error) {
	result := v1beta1.Profile{}
	err := c.restClient.
		Post().
		Resource(Profiles).
		Body(profile).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *ProfileClient) Delete(ctx context.Context, name string, opts *metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Resource(Profiles).
		Name(name).
		Body(opts).
		Do(ctx).
		Error()
}

func (c *ProfileClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1beta1.Profile, error) {
	result := v1beta1.Profile{}
	err := c.restClient.
		Get().
		Resource(Profiles).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *ProfileClient) List(ctx context.Context, opts metav1.ListOptions) (*v1beta1.ProfileList, error) {
	result := v1beta1.ProfileList{}
	err := c.restClient.
		Get().
		Resource(Profiles).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *ProfileClient) Update(ctx context.Context, profile *v1beta1.Profile) (*v1beta1.Profile, error) {
	result := v1beta1.Profile{}
	err := c.restClient.
		Put().
		Resource(Profiles).
		Body(profile).
		Do(ctx).
		Into(&result)

	return &result, err
}
