package v1alpha1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	containerv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/container/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/clientset/clientset"
	typedcontainerv1alpha1 "github.com/myeong01/ai-playground/pkg/clientset/clientset/typed/container/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/containerutils"
	"github.com/myeong01/ai-playground/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

type ContainerHandler struct {
	client typedcontainerv1alpha1.ContainerV1alpha1Interface
	logger logger.Logger
}

func NewContainerHandler(client clientset.Interface) *ContainerHandler {
	return &ContainerHandler{
		client: client.ContainerV1alpha1(),
	}
}

func (h *ContainerHandler) ListContainerInCluster(c *gin.Context) {
	containers, err := h.client.Containers("").List(c, metav1.ListOptions{})
	if err != nil {
		h.logger.Errorf("unable to list container: %v\n", err)
		c.JSON(http.StatusInternalServerError, ListContainerResponse{Error: "failed to list container in cluster wide"})
		return
	}
	containerList := make([]Container, len(containers.Items))
	for index, container := range containers.Items {
		containerList[index] = ContainerV1alpha2Container(container)
	}
	c.JSON(http.StatusOK, ListContainerResponse{Containers: containerList})
}

func (h *ContainerHandler) ListContainerInPlayground(c *gin.Context) {
	containers, err := h.client.Containers(c.Param("namespace")).List(c, metav1.ListOptions{})
	if err != nil {
		h.logger.Errorf("unable to list container: %v\n", err)
		c.JSON(http.StatusInternalServerError, ListContainerResponse{Error: "failed to list container in cluster wide"})
		return
	}
	containerList := make([]Container, len(containers.Items))
	for index, container := range containers.Items {
		containerList[index] = ContainerV1alpha2Container(container)
	}
	c.JSON(http.StatusOK, ListContainerResponse{Containers: containerList})
}

func (h *ContainerHandler) ListContainerInGroup(c *gin.Context) {
	containers, err := h.client.Containers(c.Param("namespace")).List(c, metav1.ListOptions{LabelSelector: "group=" + c.Param("group")})
	if err != nil {
		h.logger.Errorf("unable to list container: %v\n", err)
		c.JSON(http.StatusInternalServerError, ListContainerResponse{Error: "failed to list container in cluster wide"})
		return
	}
	containerList := make([]Container, len(containers.Items))
	for index, container := range containers.Items {
		containerList[index] = ContainerV1alpha2Container(container)
	}
	c.JSON(http.StatusOK, ListContainerResponse{Containers: containerList})
}

func (h *ContainerHandler) CreateContainerInGroup(c *gin.Context) {
	req := CreateContainerRequest{}
	resp := SingleContainerResponse{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	groupName := c.Param("group")
	namespace := c.Param("namespace")
	if createdContainer, err := h.client.Containers(namespace).Create(c, &containerv1alpha1.Container{
		ObjectMeta: metav1.ObjectMeta{
			Name:      groupName + "-" + req.Name,
			Namespace: namespace,
			Labels: map[string]string{
				"group": groupName,
			},
		},
		Spec: containerv1alpha1.ContainerSpec{
			ImageName:    req.ImageName,
			ResourceName: req.ResourceName,
			IsApproved:   false,
		},
	}, metav1.CreateOptions{}); err != nil {
		h.logger.Errorf("unable to create container : %v\n", err)
		resp.Error = "failed to create container"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.Container = ContainerV1alpha2Container(*createdContainer)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ContainerHandler) DeleteContainerInGroup(c *gin.Context) {
	namespace := c.Param("namespace")
	groupName := c.Param("group")
	objectName := c.Param("objectName")

	resp := DeleteContainerResponse{}

	if err := h.client.Containers(namespace).Delete(c, groupName+"-"+objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get container: %v", err)
		resp.Error = "unable to find container"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.IsSuccess = true
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ContainerHandler) GetContainerInGroup(c *gin.Context) {
	namespace := c.Param("namespace")
	groupName := c.Param("group")
	objectName := c.Param("objectName")

	resp := SingleContainerResponse{}

	if currentContainer, err := h.client.Containers(namespace).Get(c, groupName+"-"+objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get container: %v", err)
		resp.Error = "unable to find container"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.Container = ContainerV1alpha2Container(*currentContainer)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ContainerHandler) UpdateContainerInGroup(c *gin.Context) {
	namespace := c.Param("namespace")
	groupName := c.Param("group")
	objectName := c.Param("objectName")

	req := UpdateContainerRequest{}
	resp := UpdateContainerResponse{}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	containerClient := h.client.Containers(namespace)

	if currentContainer, err := containerClient.Get(c, groupName+"-"+objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get container: %v", err)
		resp.Error = "unable to find container"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		containerutils.SetStopAnnotation(&currentContainer.ObjectMeta)
		if updatedContainer, err := containerClient.Update(c, currentContainer, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to approve container: %v", err)
			resp.Error = "unable to approve container"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Start = containerutils.StopAnnotationIsSet(updatedContainer.ObjectMeta)
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ContainerHandler) ApproveContainerInGroup(c *gin.Context) {
	namespace := c.Param("namespace")
	groupName := c.Param("group")
	objectName := c.Param("objectName")

	req := ApproveContainerRequest{}
	resp := ApproveContainerResponse{}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	containerClient := h.client.Containers(namespace)

	if currentContainer, err := containerClient.Get(c, groupName+"-"+objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get container: %v", err)
		resp.Error = "unable to find container"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		currentContainer.Spec.IsApproved = req.IsApproved
		if updatedContainer, err := containerClient.Update(c, currentContainer, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to approve container: %v", err)
			resp.Error = "unable to approve container"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.IsApproved = updatedContainer.IsApproved()
		}
	}
	c.JSON(http.StatusOK, resp)
}
