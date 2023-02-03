package main

import (
	"github.com/gin-gonic/gin"
	"github.com/myeong01/ai-playground/cmd/api/apis/container/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/clientset/clientset"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	r := gin.New()

	kubeClient := clientset.NewForConfigOrDie(config.GetConfigOrDie())

	containerHandler := v1alpha1.NewContainerHandler(kubeClient)

	r.GET("/api/resource/container.ai-playground.io/v1alpha1/cluster/containers", containerHandler.ListContainerInCluster)
	r.GET("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/playground/containers", containerHandler.ListContainerInPlayground)
	r.GET("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers", containerHandler.ListContainerInGroup)
	r.POST("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers", containerHandler.CreateContainerInGroup)
	r.PUT("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/subresource/approval/:objectName", containerHandler.ApproveContainerInGroup)
	r.GET("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/object/:objectName", containerHandler.ApproveContainerInGroup)
	r.PUT("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/object/:objectName", containerHandler.UpdateContainerInGroup)
	r.DELETE("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/object/:objectName", containerHandler.DeleteContainerInGroup)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
