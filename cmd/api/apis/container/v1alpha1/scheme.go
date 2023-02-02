package v1alpha1

import (
	containerv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/container/v1alpha1"
	"strings"
)

type ListContainerResponse struct {
	Containers []Container `json:"containers,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type Container struct {
	Name       string `json:"name"`
	Group      string `json:"group,omitempty"`
	Status     string `json:"status"`
	Playground string `json:"playground"`
	ConnectUrl string `json:"connectUrl"`
}

const (
	StopAnnotation = "ai-playground-resource-stopped"
)

func ContainerV1alpha2Container(containerV1alpha1 containerv1alpha1.Container) Container {
	container := Container{
		Name:       containerV1alpha1.Name,
		Playground: strings.TrimPrefix(containerV1alpha1.Namespace, "pg-"),
	}

	if group, exist := containerV1alpha1.Labels["group"]; exist {
		container.Group = group
	}

	if containerV1alpha1.Status.AllocatedResource.Deployment.Running {
		container.Status = "Running"
		container.ConnectUrl = containerV1alpha1.Status.AllocatedResource.ConnectUrl
	} else {
		status := containerV1alpha1.Status.AllocatedResource.Deployment.Status
		if containerV1alpha1.ObjectMeta.Labels != nil {
			if _, exist := containerV1alpha1.ObjectMeta.Labels[StopAnnotation]; exist {
				status = "Stopped"
			}
		}
		container.Status = status
	}

	return container
}
