package v1alpha1

import (
	"github.com/gin-gonic/gin"
	"github.com/myeong01/ai-playground/pkg/clientset/clientset"
	containerv1alpha1 "github.com/myeong01/ai-playground/pkg/clientset/clientset/typed/container/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

type ContainerHandler struct {
	client containerv1alpha1.ContainerV1alpha1Interface
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
		h.logger.Errorf("unable to list container: %v", err)
		c.JSON(http.StatusInternalServerError, ListContainerResponse{Error: "failed to list container in cluster wide"})
		return
	}
	containerList := make([]Container, len(containers.Items))
	for index, container := range containers.Items {
		containerList[index] = ContainerV1alpha2Container(container)
	}
	c.JSON(http.StatusOK, ListContainerResponse{Containers: containerList})
}
