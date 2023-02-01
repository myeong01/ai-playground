package authorizer

import (
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
)

type Checker struct {
	ResourceAttributes *authorizationv1.ResourceAttributes
	UserInfo           *authenticationv1.UserInfo
	GroupName          string
}

func NewChecker() *Checker {
	return &Checker{
		ResourceAttributes: &authorizationv1.ResourceAttributes{},
		UserInfo:           &authenticationv1.UserInfo{},
		GroupName:          "",
	}
}

func (c *Checker) WithGroupWide(groupName string) *Checker {
	c.GroupName = groupName
	return c
}

func (c *Checker) WithVerb(verb string) *Checker {
	c.ResourceAttributes.Verb = verb
	return c
}

func (c *Checker) WithSubresource(subresource string) *Checker {
	c.ResourceAttributes.Subresource = subresource
	return c
}

func (c *Checker) WithObject(name, namespace string) *Checker {
	c.ResourceAttributes.Name = name
	c.ResourceAttributes.Namespace = namespace
	return c
}

func (c *Checker) WithObjectName(name string) *Checker {
	c.ResourceAttributes.Name = name
	return c
}

func (c *Checker) WithObjectNamespace(namespace string) *Checker {
	c.ResourceAttributes.Namespace = namespace
	return c
}

func (c *Checker) WithGVR(group, version, resource string) *Checker {
	c.ResourceAttributes.Group = group
	c.ResourceAttributes.Version = version
	c.ResourceAttributes.Resource = resource
	return c
}

func (c *Checker) WithGroup(group string) *Checker {
	c.ResourceAttributes.Group = group
	return c
}

func (c *Checker) WithVersion(version string) *Checker {
	c.ResourceAttributes.Version = version
	return c
}
func (c *Checker) WithResource(resource string) *Checker {
	c.ResourceAttributes.Resource = resource
	return c
}

func (c *Checker) WithUser(userInfo *authenticationv1.UserInfo) *Checker {
	c.UserInfo = userInfo
	return c
}

func (c *Checker) WithUsername(username string) *Checker {
	c.UserInfo.Username = username
	return c
}
