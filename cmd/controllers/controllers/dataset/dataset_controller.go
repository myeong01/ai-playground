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

package dataset

import (
	"context"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datasetv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/dataset/v1alpha1"
)

// DatasetReconciler reconciles a Dataset object
type DatasetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	imagePullSecretName = "gitlab-registry-secret"
	webImage            = "registry.gitlab.com/01ai/eng/dataman/dataman:droppy-prealpha-00.00.01-fc429cb0"
)

//+kubebuilder:rbac:groups=dataset.ai-playground.io,resources=datasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dataset.ai-playground.io,resources=datasets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dataset.ai-playground.io,resources=datasets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dataset object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *DatasetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO: add rbac permission for this logic

	dataset := &datasetv1alpha1.Dataset{}
	if err := r.Get(ctx, req.NamespacedName, dataset); err != nil {
		return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch Dataset", logger)
	}

	// subResources will be auto deleted by kubernetes with owner reference
	// nothing to do with delete dataset, just wait for dataset deletion
	if !dataset.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// TODO: create subResources [(Deployment, Service, VirtualService for web), (Deployment, Service for nfs), (PersistentVolumeClaim for storage)]

	return ctrl.Result{}, nil
}

func generateDeploymentForWeb(dataset *datasetv1alpha1.Dataset, pvcName string) *appsv1.Deployment {
	replicas := int32(1)

	// TODO fix schema for not using auth itself
	name := "dw-" + dataset.Name
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: dataset.Namespace,
			Labels: map[string]string{
				"tier":                      "web",
				"dataset-name":              dataset.Name,
				"owned-by/ai-playground.ai": "dataset-controller",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"tier":                      "web",
					"dataset-name":              dataset.Name,
					"owned-by/ai-playground.ai": "dataset-controller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"tier":                      "web",
						"deployment":                name,
						"dataset-name":              dataset.Name,
						"owned-by/ai-playground.ai": "dataset-controller",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "droppy",
							Image: webImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8989,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/files",
									Name:      "dataset",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
									ReadOnly:  false,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatasetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasetv1alpha1.Dataset{}).
		Complete(r)
}
