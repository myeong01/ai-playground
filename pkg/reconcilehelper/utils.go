package reconcilehelper

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	istioapisv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"reflect"
)

func GetErrorExceptNotFound(err error, msg string, logger logr.Logger) error {
	if err == nil || apierrs.IsNotFound(err) {
		return nil
	}
	logger.Error(err, msg)
	return err
}

func CopyDeploymentSetFields(from, to *appsv1.Deployment) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			fmt.Println("diff from label")
			requireUpdate = true
			break
		}
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			if k == "deployment.kubernetes.io/revision" {
				continue
			}
			fmt.Println("diff from annotation", k, v, from.Annotations[v])
			requireUpdate = true
			break
		}
	}
	to.Annotations = from.Annotations

	if *from.Spec.Replicas != *to.Spec.Replicas {
		fmt.Println("diff from replicas", *to.Spec.Replicas, *from.Spec.Replicas)
		to.Spec.Replicas = from.Spec.Replicas
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers, from.Spec.Template.Spec.Containers) {
		fmt.Println("diff from containers")
		requireUpdate = true
	}
	if !reflect.DeepEqual(to.Spec.Template.Spec.Volumes, from.Spec.Template.Spec.Volumes) {
		fmt.Println("------------------------------")
		fmt.Println(json.MarshalIndent(to.Spec.Template.Spec.Volumes, "", "  "))
		fmt.Println("---------------")
		fmt.Println(json.MarshalIndent(from.Spec.Template.Spec.Volumes, "", "  "))
		fmt.Println("------------------------------")
		requireUpdate = true
	}
	to.Spec.Template.Spec = from.Spec.Template.Spec

	return requireUpdate
}

// CopyServiceFields copies the owned fields from one Service to another
func CopyServiceFields(from, to *corev1.Service) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	to.Annotations = from.Annotations

	// Don't copy the entire Spec, because we can't overwrite the clusterIp field

	if !reflect.DeepEqual(to.Spec.Selector, from.Spec.Selector) {
		requireUpdate = true
	}
	to.Spec.Selector = from.Spec.Selector

	if !reflect.DeepEqual(to.Spec.Ports, from.Spec.Ports) {
		requireUpdate = true
	}
	to.Spec.Ports = from.Spec.Ports

	return requireUpdate
}

// Copy configuration related fields to another instance and returns true if there
// is a diff and thus needs to update.
func CopyVirtualService(from, to *istioapisv1beta1.VirtualService) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
			break
		}
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
			break
		}
	}
	to.Annotations = from.Annotations

	if !reflect.DeepEqual(from.Spec.Http, to.Spec.Http) {
		requireUpdate = true
	}
	to.Spec.Http = from.Spec.Http

	if !reflect.DeepEqual(from.Spec.Gateways, to.Spec.Gateways) {
		requireUpdate = true
	}
	to.Spec.Gateways = from.Spec.Gateways

	if !reflect.DeepEqual(from.Spec.Hosts, to.Spec.Hosts) {
		requireUpdate = true
	}
	to.Spec.Hosts = from.Spec.Hosts

	return requireUpdate
}
