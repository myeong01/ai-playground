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

package playground

import (
	"context"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	playgroundv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/playground/v1alpha1"
)

const (
	ChildResourceNamePrefix = "pg-"
)

// PlaygroundReconciler reconciles a Playground object
type PlaygroundReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=playground.ai-playground.io,resources=playgrounds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=playground.ai-playground.io,resources=playgrounds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=list;watch;create;update;delete
//+kubebuilder:rbac:groups=playground.ai-playground.io,resources=playgrounds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Playground object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PlaygroundReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	playground := &playgroundv1alpha1.Playground{}
	if err := r.Get(ctx, req.NamespacedName, playground); err != nil {
		return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch Playground", logger)
	}

	// subResources will be auto deleted by kubernetes with owner reference
	// nothing to do with delete playground, just wait for playground deletion
	if !playground.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if playground.Spec.IsApproved {
		namespace := generateNamespace(playground)
		if err := ctrl.SetControllerReference(playground, namespace, r.Scheme); err != nil {
			logger.Error(err, "unable to set controller reference to Namespace")
			return ctrl.Result{}, err
		}
		foundedNamespace := &corev1.Namespace{}
		err := r.Get(ctx, types.NamespacedName{Name: namespace.Name}, foundedNamespace)
		if err != nil && apierrs.IsNotFound(err) {
			logger.Info("Creating Namespace", "namespace", namespace.Name)
			err := r.Create(ctx, namespace)
			if err != nil {
				logger.Error(err, "unable to create Namespace")
				playground.Status.NamespaceName = ""
				playground.Status.NamespaceReason = err.Error()
				updateErr := r.Status().Update(ctx, playground)
				if updateErr != nil {
					logger.Error(updateErr, "failed to update Playground status")
				}
				return ctrl.Result{}, err
			}
			playground.Status.NamespaceName = namespace.Name
			playground.Status.NamespaceReason = ""
			updateErr := r.Status().Update(ctx, playground)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update Playground status")
			}
		} else if err != nil {
			logger.Error(err, "unable to fetch Namespace")
			if playground.Status.IsReady || playground.Status.NamespaceName != "" {
				playground.Status.IsReady = false
				playground.Status.NamespaceName = ""
				playground.Status.NamespaceReason = err.Error()
				err := r.Status().Update(ctx, playground)
				if err != nil {
					logger.Error(err, "unable to update Playground status")
				}
			}
			return ctrl.Result{}, err
		}

		resourceQuota := generateResourceQuota(playground)
		if err := ctrl.SetControllerReference(playground, resourceQuota, r.Scheme); err != nil {
			logger.Error(err, "unable to set controller reference to ResourceQuota")
			return ctrl.Result{}, err
		}
		foundedResourceQuota := &corev1.ResourceQuota{}
		err = r.Get(ctx, types.NamespacedName{Name: resourceQuota.Name, Namespace: resourceQuota.Namespace}, foundedResourceQuota)
		if err != nil && apierrs.IsNotFound(err) {
			logger.Info("Creating ResourceQuota", "namespace", resourceQuota.Namespace, "name", resourceQuota.Name)
			err := r.Create(ctx, resourceQuota)
			if err != nil {
				logger.Error(err, "unable to create ResourceQuota")
				playground.Status.ResourceQuotaName = ""
				playground.Status.ResourceQuotaReason = err.Error()
				updateErr := r.Status().Update(ctx, playground)
				if updateErr != nil {
					logger.Error(err, "failed to update Playground status")
				}
				return ctrl.Result{}, err
			}
			playground.Status.ResourceQuotaName = resourceQuota.Name
			playground.Status.ResourceQuotaReason = ""
			updateErr := r.Status().Update(ctx, playground)
			if updateErr != nil {
				logger.Error(err, "failed to update Playground status")
			}
		} else if err != nil {
			logger.Error(err, "error getting ResourceQuota")
			playground.Status.ResourceQuotaName = ""
			playground.Status.ResourceQuotaReason = err.Error()
			updateErr := r.Status().Update(ctx, playground)
			if updateErr != nil {
				logger.Error(err, "failed to update Playground status")
			}
			return ctrl.Result{}, err
		} else {
			if !reflect.DeepEqual(playground.Spec.ResourceQuota, foundedResourceQuota.Spec.Hard) {
				foundedResourceQuota.Spec.Hard = playground.Spec.ResourceQuota
				logger.Info("Updating ResourceQuota", "namespace", foundedResourceQuota.Namespace, "name", foundedResourceQuota.Name)
				err := r.Update(ctx, foundedResourceQuota)
				if err != nil {
					logger.Error(err, "unable to update ResourceQuota")
					playground.Status.ResourceQuotaName = ""
					playground.Status.ResourceQuotaReason = err.Error()
					updateErr := r.Status().Update(ctx, playground)
					if updateErr != nil {
						logger.Error(err, "failed to update Playground status")
					}
				}
				playground.Status.ResourceQuotaName = foundedResourceQuota.Name
				playground.Status.ResourceQuotaReason = ""
				updateErr := r.Status().Update(ctx, playground)
				if updateErr != nil {
					logger.Error(err, "failed to update Playground status")
				}
				return ctrl.Result{}, err
			}
		}
	} else {
		// TODO implement removing namespace and resource quota when playground not approved
	}

	return ctrl.Result{}, nil
}

func generateChildResourceName(parentName string) string {
	return ChildResourceNamePrefix + parentName
}

func generateNamespace(playground *playgroundv1alpha1.Playground) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateChildResourceName(playground.Name),
		},
	}
}

func generateResourceQuota(playground *playgroundv1alpha1.Playground) *corev1.ResourceQuota {
	name := generateChildResourceName(playground.Name)
	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: playground.Spec.ResourceQuota,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlaygroundReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&playgroundv1alpha1.Playground{}).
		Complete(r)
}
