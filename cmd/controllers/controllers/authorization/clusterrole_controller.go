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

package authorization

import (
	"context"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
)

const (
	ChildResourceNamePrefix = "ac-"
	LabelKey                = "parent"
	LabelPrefixClusterRole  = "clusterrole/"
	LabelPrefixRole         = "role/"
)

// ClusterRoleReconciler reconciles a ClusterRole object
type ClusterRoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=clusterroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=clusterroles/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClusterRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	clusterRole := &authorizationv1alpha1.ClusterRole{}
	if err := r.Get(ctx, req.NamespacedName, clusterRole); err != nil {
		return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch ClusterRole", logger)
	}

	if !clusterRole.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// TODO prevent Creating with unapproved parent ClusterRole with Webhook
	if clusterRole.Spec.ParentClusterRoleName != "" {
		parentClusterRole := &authorizationv1alpha1.ClusterRole{}
		if err := r.Get(ctx, types.NamespacedName{Name: clusterRole.Spec.ParentClusterRoleName}, parentClusterRole); err != nil {
			clusterRole.Status.IsFailed = true
			clusterRole.Status.Reason = err.Error()
			return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch parent ClusterRole", logger)
		}
		if len(clusterRole.OwnerReferences) == 0 {
			if err := ctrl.SetControllerReference(parentClusterRole, clusterRole, r.Scheme); err != nil {
				logger.Error(err, "unable to set controller reference to ClusterRole")
				return ctrl.Result{}, err
			}
			if err := r.Update(ctx, clusterRole); err != nil {
				logger.Error(err, "unable to update ClusterRole OwnerReference")
				return ctrl.Result{}, err
			}
		}
		if clusterRole.ObjectMeta.Labels[LabelKey] != LabelPrefixClusterRole+clusterRole.Spec.ParentClusterRoleName {
			clusterRole.ObjectMeta.Labels[LabelKey] = LabelPrefixClusterRole + clusterRole.Spec.ParentClusterRoleName
			if err := r.Update(ctx, clusterRole); err != nil {
				logger.Error(err, "unable to update ClusterRole label")
				return ctrl.Result{}, err
			}
		}
		if updatedClusterRoleRule, updated := r.removeChildRuleNotInParent(parentClusterRole.Spec.Rules, clusterRole.Spec.Rules); updated {
			clusterRole.Spec.Rules = updatedClusterRoleRule
			if err := r.Update(ctx, clusterRole); err != nil {
				logger.Error(err, "unable to update ClusterRole Spec")
				return ctrl.Result{}, err
			}
		}
	}

	if !clusterRole.Spec.IsApproved {
		return ctrl.Result{}, nil
	}

	kubeClusterRole := generateKubeClusterRole(clusterRole)
	foundedClusterRole := &rbacv1.ClusterRole{}
	err := r.Get(ctx, types.NamespacedName{Name: kubeClusterRole.Name}, foundedClusterRole)
	if err != nil && apierrs.IsNotFound(err) {
		logger.Info("Creating ClusterRole", "name", clusterRole.Name)
		err := r.Create(ctx, clusterRole)
		if err != nil {
			logger.Error(err, "unable to create ClusterRole")
			clusterRole.Status.IsFailed = true
			clusterRole.Status.Reason = err.Error()
			updateErr := r.Status().Update(ctx, clusterRole)
			if updateErr != nil {
				logger.Error(err, "unable to update ClusterRole status")
			}
			return ctrl.Result{}, err
		}
		clusterRole.Status = authorizationv1alpha1.ClusterRoleStatus{
			ClusterRoleName: clusterRole.Name,
			IsFailed:        false,
			Reason:          "",
			Rules:           clusterRole.Spec.Rules,
		}
		updateErr := r.Status().Update(ctx, clusterRole)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update ClusterRole status")
			return ctrl.Result{}, updateErr
		}
	} else if err != nil {
		logger.Error(err, "error getting ClusterRole")
		clusterRole.Status.IsFailed = true
		clusterRole.Status.Reason = err.Error()
		updateErr := r.Status().Update(ctx, clusterRole)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update ClusterRole status")
		}
		return ctrl.Result{}, err
	} else {
		if !reflect.DeepEqual(kubeClusterRole.Rules, clusterRole.Spec.Rules) {
			kubeClusterRole.Rules = clusterRole.Spec.Rules
			updateErr := r.Update(ctx, kubeClusterRole)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update KubeClusterRole spec")
				return ctrl.Result{}, updateErr
			}
			clusterRole.Status.Rules = clusterRole.Spec.Rules
			clusterRole.Status.IsChildChecked = false
			updateErr = r.Status().Update(ctx, clusterRole)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update ClusterRole status")
				return ctrl.Result{}, updateErr
			}
		} else if !reflect.DeepEqual(clusterRole.Spec.Rules, clusterRole.Status.Rules) {
			clusterRole.Status.Rules = clusterRole.Spec.Rules
			updateErr := r.Status().Update(ctx, clusterRole)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update ClusterRole status")
				return ctrl.Result{}, updateErr
			}
		}
		if !clusterRole.Status.IsChildChecked {
			childClusterRoleList, err := r.listAllChildClusterRole(ctx, clusterRole)
			if err != nil {
				logger.Error(err, "failed to list child ClusterRole")
				return ctrl.Result{}, err
			}
			for _, childClusterRole := range childClusterRoleList.Items {
				updatedChildClusterRoleRule, updated := r.removeChildRuleNotInParent(clusterRole.Spec.Rules, childClusterRole.Spec.Rules)
				if updated {
					childClusterRole.Spec.Rules = updatedChildClusterRoleRule
					if err := r.Update(ctx, &childClusterRole); err != nil {
						logger.Error(err, "failed to update childClusterRole spec")
						return ctrl.Result{}, err
					}
				}
			}

			childRoleList, err := r.listAllChildRole(ctx, clusterRole)
			if err != nil {
				logger.Error(err, "failed to list child Role")
				return ctrl.Result{}, err
			}
			for _, childRole := range childRoleList.Items {
				updatedChildRoleRule, updated := r.removeChildRuleNotInParent(clusterRole.Spec.Rules, childRole.Spec.Rules)
				if updated {
					childRole.Spec.Rules = updatedChildRoleRule
					if err := r.Update(ctx, &childRole); err != nil {
						logger.Error(err, "failed to update childRole spec")
						return ctrl.Result{}, err
					}
				}
			}

			clusterRole.Status.IsChildChecked = true
			updateErr := r.Status().Update(ctx, clusterRole)
			if updateErr != nil {
				logger.Error(updateErr, "failed to update ClusterRole status")
				return ctrl.Result{}, updateErr
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterRoleReconciler) removeChildRuleNotInParent(parentRules, childRules []rbacv1.PolicyRule) ([]rbacv1.PolicyRule, bool) {
	if parentRules == nil || childRules == nil {
		return nil, childRules != nil
	}
	newChildRules := make([]rbacv1.PolicyRule, 0)
	isChildRuleNeedUpdate := false
	for _, childRule := range childRules {
		for _, parentRule := range parentRules {
			if reflect.DeepEqual(parentRule, childRule) {
				newChildRules = append(newChildRules, childRule)
				continue
			}
		}
		isChildRuleNeedUpdate = true
	}
	return newChildRules, isChildRuleNeedUpdate
}

func (r *ClusterRoleReconciler) listAllChildClusterRole(ctx context.Context, clusterRole *authorizationv1alpha1.ClusterRole) (*authorizationv1alpha1.ClusterRoleList, error) {
	clusterRoleList := &authorizationv1alpha1.ClusterRoleList{}
	if err := r.List(ctx, clusterRoleList, &client.MatchingLabels{
		LabelKey: LabelPrefixClusterRole + clusterRole.Name,
	}); err != nil {
		return nil, err
	}
	return clusterRoleList, nil
}

func (r *ClusterRoleReconciler) listAllChildRole(ctx context.Context, clusterRole *authorizationv1alpha1.ClusterRole) (*authorizationv1alpha1.RoleList, error) {
	// TODO check range of namespace
	roleList := &authorizationv1alpha1.RoleList{}
	if err := r.List(ctx, roleList, &client.MatchingLabels{
		LabelKey: LabelPrefixClusterRole + clusterRole.Name,
	}); err != nil {
		return nil, err
	}
	return roleList, nil
}

func generateChildResourceName(name string) string {
	return ChildResourceNamePrefix + name
}

func generateKubeClusterRole(clusterRole *authorizationv1alpha1.ClusterRole) *rbacv1.ClusterRole {
	name := generateChildResourceName(clusterRole.Name)
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: clusterRole.Spec.Rules,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authorizationv1alpha1.ClusterRole{}).
		Complete(r)
}
