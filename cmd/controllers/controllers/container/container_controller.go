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

package container

import (
	"context"
	containerv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/container/v1alpha1"
	imagev1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/image/v1alpha1"
	resourcev1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/resource/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper"
	istioapiv1beta1 "istio.io/api/networking/v1beta1"
	istioapisv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const (
	StopAnnotation          = "ai-playground-resource-stopped"
	LastActivityAnnotation  = "container.ai-playground.ai/last_activity"
	PrefixEnvVar            = "URL_PREFIX"
	GatewayName             = "default/main-gateway"
	HostName                = "myeongsuk.ml"
	ChildResourceNamePrefix = "cu-"
)

const (
	DefaultPort = 80
)

// ContainerReconciler reconciles a Container object
type ContainerReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=container.ai-playground.io,resources=containers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=list;watch;create;update;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=list;watch;create;update;delete
//+kubebuilder:rbac:groups=container.ai-playground.io,resources=containers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=container.ai-playground.io,resources=containers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Container object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
// TODO all Update function called must be changed to Patch function
func (r *ContainerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	container := &containerv1alpha1.Container{}
	if err := r.Get(ctx, req.NamespacedName, container); err != nil {
		return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch Container", logger)
	}

	// subResources will be auto deleted by kubernetes with owner reference
	// nothing to do with delete container, just wait for container deletion
	if !container.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if container.Spec.IsApproved {
		return r.ApproveContainerReconcile(ctx, container, req)
	} else {
		return r.UnApproveContainerReconcile(ctx, container, req)
	}
}

func (r *ContainerReconciler) ApproveContainerReconcile(ctx context.Context, container *containerv1alpha1.Container, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get Resource and Image for generate Deployment
	resource := &resourcev1alpha1.Resource{}
	if err := r.Get(ctx, types.NamespacedName{Name: container.Spec.ResourceName, Namespace: container.Namespace}, resource); err != nil {
		logger.Error(err, "unable to fetch resource for container")
		return ctrl.Result{}, err
	}
	image := &imagev1alpha1.Image{}
	if err := r.Get(ctx, types.NamespacedName{Name: container.Spec.ImageName, Namespace: container.Namespace}, image); err != nil {
		logger.Error(err, "unable to fetch image for container")
		return ctrl.Result{}, err
	}

	// reconcile deployment
	deployment := generateDeployment(container, resource, image)
	if err := ctrl.SetControllerReference(container, deployment, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference to Deployment")
		return ctrl.Result{}, err
	}
	foundDeployment := &appsv1.Deployment{}
	justCreated := false
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && apierrs.IsNotFound(err) {
		logger.Info("Creating Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
		err := r.Create(ctx, deployment)
		justCreated = true
		if err != nil {
			logger.Error(err, "unable to create Deployment")
			container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
				Created: false,
				Running: false,
				Failed:  true,
				Name:    deployment.Name,
				Reason:  err.Error(),
				Status:  "Failed to create Deployment",
			}
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
			return ctrl.Result{}, err
		}
		container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
			Created: true,
			Running: false,
			Failed:  false,
			Name:    deployment.Name,
			Status:  "Deployment success created",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	} else if err != nil {
		logger.Error(err, "error getting Deployment")
		container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  true,
			Name:    deployment.Name,
			Reason:  err.Error(),
			Status:  "Failed to lookup Deployment",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
		return ctrl.Result{}, err
	}

	if !justCreated && reconcilehelper.CopyDeploymentSetFields(deployment, foundDeployment) {
		logger.Info("Updating Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
		foundDeployment.SetResourceVersion("")
		err = r.Update(ctx, foundDeployment)
		if err != nil {
			logger.Error(err, "unable to update Deployment")
			container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
				Created: false,
				Running: false,
				Failed:  true,
				Name:    deployment.Name,
				Reason:  err.Error(),
				Status:  "Failed to update Deployment",
			}
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
			return ctrl.Result{}, err
		}
	}

	// reconcile service
	svc := generateService(container)
	if err := ctrl.SetControllerReference(container, svc, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference to Service")
		return ctrl.Result{}, err
	}
	foundSvc := &corev1.Service{}
	justCreated = false
	err = r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)
	if err != nil && apierrs.IsNotFound(err) {
		logger.Info("Creating Service", "namespace", svc.Namespace, "name", svc.Name)
		err = r.Create(ctx, svc)
		justCreated = true
		if err != nil {
			logger.Error(err, "unable to create Service")
			container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
				Created: false,
				Running: false,
				Failed:  true,
				Name:    svc.Name,
				Reason:  err.Error(),
				Status:  "Failed to create Service",
			}
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
			return ctrl.Result{}, err
		}
		container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
			Created: true,
			Running: true,
			Failed:  false,
			Name:    svc.Name,
			Status:  "Service success created",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	} else if err != nil {
		logger.Error(err, "error getting Service")
		container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  true,
			Name:    svc.Name,
			Reason:  err.Error(),
			Status:  "Failed to lookup Service",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
		return ctrl.Result{}, err
	}

	if !justCreated && reconcilehelper.CopyServiceFields(svc, foundSvc) {
		logger.Info("Updating Service", "namespace", svc.Namespace, "name", svc.Name)
		err = r.Update(ctx, foundSvc)
		if err != nil {
			logger.Error(err, "unable to update Service")
			container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
				Created: false,
				Running: false,
				Failed:  true,
				Name:    svc.Name,
				Reason:  err.Error(),
				Status:  "Failed to update Service",
			}
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
			return ctrl.Result{}, err
		}
	}

	// reconcile virtual service
	vs := generateVirtualService(container)
	if err := ctrl.SetControllerReference(container, vs, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference to VirtualService")
		return ctrl.Result{}, err
	}
	foundVs := &istioapisv1beta1.VirtualService{}
	justCreated = false
	err = r.Get(ctx, types.NamespacedName{Name: vs.Name, Namespace: vs.Namespace}, foundVs)
	if err != nil && apierrs.IsNotFound(err) {
		logger.Info("Creating VirtualService", "namespace", vs.Namespace, "name", vs.Name)
		err = r.Create(ctx, vs)
		justCreated = true
		if err != nil {
			logger.Error(err, "unable to create VirtualService")
			container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
				Created: false,
				Running: false,
				Failed:  true,
				Name:    vs.Name,
				Reason:  err.Error(),
				Status:  "Failed to create VirtualService",
			}
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
			return ctrl.Result{}, err
		}
		container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
			Created: true,
			Running: true,
			Failed:  false,
			Name:    vs.Name,
			Status:  "VirtualService success created",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	} else if err != nil {
		logger.Error(err, "error getting VirtualService")
		container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  true,
			Name:    vs.Name,
			Reason:  err.Error(),
			Status:  "Failed to lookup VirtualService",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
		return ctrl.Result{}, err
	}

	if !justCreated && reconcilehelper.CopyVirtualService(vs, foundVs) {
		logger.Info("Updating VirtualService", "namespace", vs.Namespace, "name", vs.Name)
		err = r.Update(ctx, foundVs)
		if err != nil {
			logger.Error(err, "unable to update VirtualService")
			container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
				Created: false,
				Running: false,
				Failed:  true,
				Name:    vs.Name,
				Reason:  err.Error(),
				Status:  "Failed to update VirtualService",
			}
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
			return ctrl.Result{}, err
		}
	}

	// sync Deployment status
	curDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, curDeployment)
	if err != nil {
		logger.Error(err, "unable to lookup Deployment")
		container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  true,
			Name:    curDeployment.Name,
			Reason:  err.Error(),
			Status:  "Failed to lookup Deployment",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
		return ctrl.Result{}, err
	}

	if curDeployment.Status.AvailableReplicas == 0 && container.Status.AllocatedResource.Deployment.Running == true {
		container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
			Created: true,
			Running: false,
			Failed:  false,
			Name:    curDeployment.Name,
			Status:  "Pod recreating",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	} else if curDeployment.Status.AvailableReplicas == 1 && container.Status.AllocatedResource.Deployment.Running == false {
		container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
			Created: true,
			Running: true,
			Failed:  false,
			Name:    curDeployment.Name,
			Status:  "Pod running",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	}
	connectUrl := "https://" + HostName + getUrlPrefix(container)
	if container.Status.AllocatedResource.ConnectUrl != connectUrl {
		container.Status.AllocatedResource.ConnectUrl = connectUrl
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	}
	return ctrl.Result{}, nil
}

func (r *ContainerReconciler) UnApproveContainerReconcile(ctx context.Context, container *containerv1alpha1.Container, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	childResourceName := generateChildResourceName(container.Name)

	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: childResourceName, Namespace: container.Namespace}, deployment); err == nil {
		deleteErr := r.Delete(ctx, deployment)
		needUpdate := false
		if deleteErr == nil || apierrs.IsNotFound(err) {
			if container.Status.AllocatedResource.Deployment.Status != "ready for approve" {
				container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
					Created: false,
					Running: false,
					Failed:  false,
					Name:    "",
					Status:  "ready for approve",
				}
				needUpdate = true
			}
		} else {
			container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
				Created: true,
				Running: false,
				Failed:  false,
				Name:    childResourceName,
				Status:  "failed to delete",
				Reason:  deleteErr.Error(),
			}
			needUpdate = true
		}
		if needUpdate {
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
		}
	} else if apierrs.IsNotFound(err) && container.Status.AllocatedResource.Deployment.Status != "ready for approve" {
		container.Status.AllocatedResource.Deployment = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  false,
			Name:    "",
			Status:  "ready for approve",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	}

	service := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: childResourceName, Namespace: container.Namespace}, service); err == nil {
		deleteErr := r.Delete(ctx, service)
		needUpdate := false
		if deleteErr == nil || apierrs.IsNotFound(err) {
			if container.Status.AllocatedResource.Service.Status != "ready for approve" {
				container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
					Created: false,
					Running: false,
					Failed:  false,
					Name:    "",
					Status:  "ready for approve",
				}
				needUpdate = true
			}
		} else {
			container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
				Created: true,
				Running: false,
				Failed:  false,
				Name:    childResourceName,
				Status:  "failed to delete",
				Reason:  deleteErr.Error(),
			}
			needUpdate = true
		}
		if needUpdate {
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
		}
	} else if apierrs.IsNotFound(err) && container.Status.AllocatedResource.Service.Status != "ready for approve" {
		container.Status.AllocatedResource.Service = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  false,
			Name:    "",
			Status:  "ready for approve",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	}

	virtualService := &istioapisv1beta1.VirtualService{}
	if err := r.Get(ctx, types.NamespacedName{Name: childResourceName, Namespace: container.Namespace}, virtualService); err == nil {
		deleteErr := r.Delete(ctx, virtualService)
		needUpdate := false
		if deleteErr == nil || apierrs.IsNotFound(err) {
			if container.Status.AllocatedResource.VirtualService.Status != "ready for approve" {
				container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
					Created: false,
					Running: false,
					Failed:  false,
					Name:    "",
					Status:  "ready for approve",
				}
				needUpdate = true
			}
		} else {
			container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
				Created: true,
				Running: false,
				Failed:  false,
				Name:    childResourceName,
				Status:  "failed to delete",
				Reason:  deleteErr.Error(),
			}
			needUpdate = true
		}
		if needUpdate {
			updateErr := r.Status().Update(ctx, container)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update container status")
			}
		}
	} else if apierrs.IsNotFound(err) && container.Status.AllocatedResource.VirtualService.Status != "ready for approve" {
		container.Status.AllocatedResource.VirtualService = containerv1alpha1.ResourceStatus{
			Created: false,
			Running: false,
			Failed:  false,
			Name:    "",
			Status:  "ready for approve",
		}
		updateErr := r.Status().Update(ctx, container)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update container status")
		}
	}

	return ctrl.Result{}, nil
}

func SetStopAnnotation(meta *metav1.ObjectMeta) {
	if meta == nil {
		return
	}
	t := time.Now()
	if meta.GetAnnotations() != nil {
		meta.Annotations[StopAnnotation] = t.Format(t.Format(time.RFC3339))
	} else {
		meta.SetAnnotations(map[string]string{
			StopAnnotation: t.Format(time.RFC3339),
		})
	}

	if meta.GetAnnotations() != nil {
		if _, ok := meta.GetAnnotations()[LastActivityAnnotation]; ok {
			delete(meta.GetAnnotations(), LastActivityAnnotation)
		}
	}
}

func StopAnnotationIsSet(meta metav1.ObjectMeta) bool {
	if meta.GetAnnotations() == nil {
		return false
	}

	_, ok := meta.GetAnnotations()[StopAnnotation]
	return ok
}

func generateChildResourceName(parentName string) string {
	return ChildResourceNamePrefix + parentName
}

func generateDeployment(container *containerv1alpha1.Container, resource *resourcev1alpha1.Resource, image *imagev1alpha1.Image) *appsv1.Deployment {
	replicas := int32(1)
	if StopAnnotationIsSet(container.ObjectMeta) {
		replicas = 0
	}

	name := generateChildResourceName(container.Name)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: container.Namespace,
			Labels: map[string]string{
				"container-name":            container.Name,
				"owned-by/ai-playground.ai": "container-controller",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"deployment":                name,
					"owned-by/ai-playground.ai": "container-controller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deployment":                       name,
						"container-name":                   container.Name,
						"owned-by/ai-playground.ai":        "container-controller",
						"sidecar.dataset.ai-playground.io": "true",
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: stringSliceToLocalObjectReferenceSlice(image.Spec.ImagePullSecrets),
					Containers: []corev1.Container{
						{
							Name:                     "container",
							Image:                    image.Spec.ImageNameWithTag,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							Resources:                corev1.ResourceRequirements{},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: int32(container.Spec.Port),
								},
							},
							Env: []corev1.EnvVar{
								envVar(PrefixEnvVar, getUrlPrefix(container)),
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func generateService(container *containerv1alpha1.Container) *corev1.Service {
	name := generateChildResourceName(container.Name)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: container.Namespace,
			Labels: map[string]string{
				"container-name":            container.Name,
				"owned-by/ai-playground.ai": "container-controller",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"deployment": name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http-" + name,
					Protocol:   corev1.ProtocolTCP,
					Port:       DefaultPort,
					TargetPort: intstr.FromInt(int(container.Spec.Port)),
				},
			},
		},
	}
}

func getUrlPrefix(container *containerv1alpha1.Container) string {
	return "/notebook/" + container.GetNamespace() + "/" + container.GetName()
}

func generateVirtualService(container *containerv1alpha1.Container) *istioapisv1beta1.VirtualService {
	name := generateChildResourceName(container.Name)
	return &istioapisv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: container.Namespace,
			Labels: map[string]string{
				"container-name":            container.Name,
				"owned-by/ai-playground.ai": "container-controller",
			},
		},
		Spec: istioapiv1beta1.VirtualService{
			Gateways: []string{GatewayName},
			Hosts:    []string{HostName},
			Http: []*istioapiv1beta1.HTTPRoute{
				{
					Match: []*istioapiv1beta1.HTTPMatchRequest{
						{
							Uri: &istioapiv1beta1.StringMatch{
								MatchType: &istioapiv1beta1.StringMatch_Prefix{Prefix: getUrlPrefix(container)},
							},
						},
					},
					Route: []*istioapiv1beta1.HTTPRouteDestination{
						{
							Destination: &istioapiv1beta1.Destination{
								Host: name,
								Port: &istioapiv1beta1.PortSelector{
									Number: DefaultPort,
								},
							},
						},
					},
				},
			},
		},
	}
}

func envVar(key, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  key,
		Value: value,
	}
}

func stringSliceToLocalObjectReferenceSlice(strs []string) []corev1.LocalObjectReference {
	if strs == nil {
		return nil
	}
	refs := make([]corev1.LocalObjectReference, len(strs))
	for index, str := range strs {
		refs[index] = corev1.LocalObjectReference{
			Name: str,
		}
	}
	return refs
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContainerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&containerv1alpha1.Container{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&istioapisv1beta1.VirtualService{}).
		Complete(r)
}
