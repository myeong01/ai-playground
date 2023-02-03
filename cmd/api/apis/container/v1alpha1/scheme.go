package v1alpha1

import (
	containerv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/container/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/containerutils"
	"strings"
)

type ListContainerResponse struct {
	Containers []Container `json:"containers,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type CreateContainerRequest struct {
	Name         string `json:"name"`
	ImageName    string `json:"imageName"`
	ResourceName string `json:"resourceName"`
}

type SingleContainerResponse struct {
	Container Container `json:"container,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type DeleteContainerResponse struct {
	IsSuccess bool   `json:"isSuccess"`
	Error     string `json:"error,omitempty"`
}

type ApproveContainerResponse struct {
	IsApproved bool   `json:"isApproved"`
	Error      string `json:"error,omitempty"`
}

type ApproveContainerRequest struct {
	IsApproved bool `json:"isApproved"`
}

type UpdateContainerRequest struct {
	Start bool `json:"start"`
}

type UpdateContainerResponse struct {
	Start bool   `json:"start"`
	Error string `json:"error"`
}

type Container struct {
	Name         string `json:"name"`
	Group        string `json:"group,omitempty"`
	Status       string `json:"status"`
	Playground   string `json:"playground"`
	ImageName    string `json:"imageName"`
	ResourceName string `json:"resourceName"`
	ConnectUrl   string `json:"connectUrl"`
}

func ContainerV1alpha2Container(containerV1alpha1 containerv1alpha1.Container) Container {
	container := Container{
		Name:         containerV1alpha1.Name,
		Playground:   strings.TrimPrefix(containerV1alpha1.Namespace, "pg-"),
		ImageName:    containerV1alpha1.Spec.ImageName,
		ResourceName: containerV1alpha1.Spec.ResourceName,
	}

	if group, exist := containerV1alpha1.Labels["group"]; exist {
		container.Group = group
		container.Name = strings.TrimPrefix(container.Name, container.Group+"-")
	}

	if containerV1alpha1.Status.AllocatedResource.Deployment.Running {
		container.Status = "Running"
		container.ConnectUrl = containerV1alpha1.Status.AllocatedResource.ConnectUrl
	} else {
		status := containerV1alpha1.Status.AllocatedResource.Deployment.Status
		if containerutils.StopAnnotationIsSet(containerV1alpha1.ObjectMeta) {
			status = "Stopped"
		}
		container.Status = status
	}

	return container
}
