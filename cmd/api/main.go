package main

import (
	"github.com/gin-gonic/gin"
	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/api/apis/authorization/v1alpha1"
	containerv1alpha1 "github.com/myeong01/ai-playground/cmd/api/apis/container/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/clientset/clientset"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	r := gin.New()

	kubeClient := clientset.NewForConfigOrDie(config.GetConfigOrDie())

	containerHandler := containerv1alpha1.NewContainerHandler(kubeClient)
	groupHandler := authorizationv1alpha1.NewGroupHandler(kubeClient)
	roleHandler := authorizationv1alpha1.NewRoleHandler(kubeClient)

	r.GET("/api/resource/container.ai-playground.io/v1alpha1/cluster/containers", containerHandler.ListContainerInCluster)
	r.GET("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/playground/containers", containerHandler.ListContainerInPlayground)
	r.GET("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers", containerHandler.ListContainerInGroup)
	r.POST("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers", containerHandler.CreateContainerInGroup)
	r.PUT("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/subresource/approval/:objectName", containerHandler.ApproveContainerInGroup)
	r.GET("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/object/:objectName", containerHandler.ApproveContainerInGroup)
	r.PUT("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/object/:objectName", containerHandler.UpdateContainerInGroup)
	r.DELETE("/api/resource/container.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/containers/object/:objectName", containerHandler.DeleteContainerInGroup)

	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/cluster/groups", groupHandler.ListGroupInCluster)
	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/groups", groupHandler.ListGroupInPlayground)
	r.POST("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/groups", groupHandler.CreateGroupInPlayground)
	r.DELETE("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/groups/object/:objectName", groupHandler.DeleteGroupInPlayground)
	r.PUT("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/groups/object/:objectName", groupHandler.UpdateGroupInPlayground)
	r.PUT("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/groups/subresource/approval/:objectName", groupHandler.ApproveGroupInPlayground)
	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/groups/object/:objectName", groupHandler.GetGroupInPlayground)

	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/cluster/roles", roleHandler.ListRoleInCluster)
	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/roles", roleHandler.ListRoleInPlayground)
	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/roles", roleHandler.ListRoleInGroup)
	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/roles/object/:objectName", roleHandler.GetRoleInPlayground)
	r.GET("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/roles/object/:objectName", roleHandler.GetRoleInGroup)
	r.POST("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/roles", roleHandler.CreateRoleInPlayground)
	r.POST("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/roles", roleHandler.CreateRoleInGroup)
	r.PUT("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/roles/object/:objectName", roleHandler.UpdateRoleInPlayground)
	r.PUT("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/roles/object/:objectName", roleHandler.UpdateRoleInGroup)
	r.DELETE("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/roles/object/:objectName", roleHandler.DeleteRoleInPlayground)
	r.DELETE("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/roles/object/:objectName", roleHandler.DeleteRoleInGroup)
	r.PUT("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/playground/roles/subresource/approval/:objectName", roleHandler.ApproveRoleInPlayground)
	r.PUT("/api/resource/authorization.ai-playground.io/v1alpha1/namespace/:namespace/group/:group/roles/subresource/approval/:objectName", roleHandler.ApproveRoleInGroup)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
