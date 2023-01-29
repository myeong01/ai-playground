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
	"errors"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
)

// RoleReconciler reconciles a Role object
type RoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=roles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=roles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Role object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	role := &authorizationv1alpha1.Role{}
	if err := r.Get(ctx, req.NamespacedName, role); err != nil {
		return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch Role", logger)
	}

	if !role.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if role.Spec.ParentRole != nil {
		var parentRoleObject metav1.Object
		var parentRoleRules []rbacv1.PolicyRule
		var labelPrefix string
		switch role.Spec.ParentRole.Type {
		case authorizationv1alpha1.TypeClusterRole:
			parentClusterRole := &authorizationv1alpha1.ClusterRole{}
			if err := r.Get(ctx, types.NamespacedName{Name: role.Spec.ParentRole.Name}, parentClusterRole); err != nil {
				logger.Error(err, "unable to fetch parent ClusterRole")
				role.Status.IsFailed = true
				role.Status.Reason = err.Error()
				if err := r.Status().Update(ctx, role); err != nil {
					logger.Error(err, "failed to update Role status")
				}
				return ctrl.Result{}, err
			}
			parentRoleObject = parentClusterRole
			parentRoleRules = parentClusterRole.Spec.Rules
			labelPrefix = ClusterRoleChildResourceNamePrefix
		case authorizationv1alpha1.TypeRole:
			parentRole := &authorizationv1alpha1.Role{}
			if err := r.Get(ctx, types.NamespacedName{Name: role.Spec.ParentRole.Name, Namespace: role.Namespace}, parentRole); err != nil {
				logger.Error(err, "unable to fetch parent Role")
				role.Status.IsFailed = true
				role.Status.Reason = err.Error()
				if err := r.Status().Update(ctx, role); err != nil {
					logger.Error(err, "failed to update Role status")
				}
				return ctrl.Result{}, err
			}
			parentRoleObject = parentRole
			parentRoleRules = parentRole.Spec.Rules
			labelPrefix = RoleChildResourceNamePrefix
		default:
			logger.Error(errors.New("unknown type"), "unknown parent role type")
			role.Status.IsFailed = true
			role.Status.Reason = "unknown parent rolt type [" + role.Spec.ParentRole.Type + "]"
			if err := r.Update(ctx, role); err != nil {
				logger.Error(err, "failed to update Role status")
			}
			return ctrl.Result{}, nil
		}
		if len(role.OwnerReferences) == 0 {
			if err := ctrl.SetControllerReference(parentRoleObject, role, r.Scheme); err != nil {
				logger.Error(err, "unable to set controller reference to Role")
				return ctrl.Result{}, err
			}
			if err := r.Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role OwnerReference")
				return ctrl.Result{}, err
			}
		}
		if role.ObjectMeta.Labels[LabelKey] != labelPrefix+role.Spec.ParentRole.Name {
			role.ObjectMeta.Labels[LabelKey] = labelPrefix + role.Spec.ParentRole.Name
			if err := r.Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role label")
				return ctrl.Result{}, err
			}
		}
		if updatedRoleRule, updated := r.removeChildRuleNotInParent(parentRoleRules, role.Spec.Rules); updated {
			role.Spec.Rules = updatedRoleRule
			if err := r.Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role Spec")
				return ctrl.Result{}, err
			}
		}
	}

	if !role.Spec.IsApproved {
		return ctrl.Result{}, nil
	}

	kubeRole := generateKubeRole(role)
	if err := ctrl.SetControllerReference(role, kubeRole, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference to KubeRole")
		return ctrl.Result{}, err
	}
	foundedKubeRole := &rbacv1.Role{}
	if err := r.Get(ctx, types.NamespacedName{Name: kubeRole.Name, Namespace: kubeRole.Namespace}, foundedKubeRole); err != nil && apierrs.IsNotFound(err) {
		logger.Info("Creating Role", "name", kubeRole.Name, "namespace", kubeRole.Namespace)
		if err := r.Create(ctx, kubeRole); err != nil {
			logger.Error(err, "unable to create Role")
			role.Status.IsFailed = true
			role.Status.Reason = err.Error()
			if err := r.Status().Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role status")
			}
			return ctrl.Result{}, err
		}
		role.Status = authorizationv1alpha1.RoleStatus{
			RoleName: kubeRole.Name,
			IsFailed: false,
			Reason:   "",
			Rules:    role.Spec.Rules,
		}
		if err := r.Status().Update(ctx, role); err != nil {
			logger.Error(err, "unable to update Role status")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		logger.Error(err, "error getting KubeRole")
		role.Status.IsFailed = true
		role.Status.Reason = err.Error()
		if err := r.Status().Update(ctx, role); err != nil {
			logger.Error(err, "unable to update Role status")
		}
		return ctrl.Result{}, err
	} else {
		if !reflect.DeepEqual(foundedKubeRole.Rules, role.Spec.Rules) {
			foundedKubeRole.Rules = role.Spec.Rules
			if err := r.Update(ctx, foundedKubeRole); err != nil {
				logger.Error(err, "unable to update KubeRole spec")
				return ctrl.Result{}, err
			}
			role.Status.Rules = foundedKubeRole.Rules
			role.Status.IsChildChecked = false
			if err := r.Status().Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role status")
				return ctrl.Result{}, err
			}
		} else if !reflect.DeepEqual(role.Spec.Rules, role.Status.Rules) {
			role.Status.Rules = role.Spec.Rules
			if err := r.Status().Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role status")
				return ctrl.Result{}, err
			}
		}
		if !role.Status.IsChildChecked {
			if childRoleList, err := r.listAllChildRole(ctx, role); err != nil {
				logger.Error(err, "unable to list child Role")
				return ctrl.Result{}, err
			} else {
				for _, childRole := range childRoleList.Items {
					if updatedChildRoleRule, updated := r.removeChildRuleNotInParent(role.Spec.Rules, childRole.Spec.Rules); updated {
						childRole.Spec.Rules = updatedChildRoleRule
						if err := r.Update(ctx, &childRole); err != nil {
							logger.Error(err, "unable to update childRole spec")
							return ctrl.Result{}, err
						}
					}
				}
			}

			role.Status.IsChildChecked = true
			if err := r.Status().Update(ctx, role); err != nil {
				logger.Error(err, "unable to update Role status")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *RoleReconciler) removeChildRuleNotInParent(parentRules, childRules []rbacv1.PolicyRule) ([]rbacv1.PolicyRule, bool) {
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

func (r *RoleReconciler) listAllChildRole(ctx context.Context, role *authorizationv1alpha1.Role) (*authorizationv1alpha1.RoleList, error) {
	// TODO check range of namespace
	roleList := &authorizationv1alpha1.RoleList{}
	if err := r.List(ctx, roleList, &client.MatchingLabels{
		LabelKey: LabelPrefixRole + role.Name,
	}); err != nil {
		return nil, err
	}
	return roleList, nil
}

func generateKubeRole(role *authorizationv1alpha1.Role) *rbacv1.Role {
	name := RoleChildResourceNamePrefix + role.Name
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: role.Namespace,
		},
		Rules: role.Spec.Rules,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authorizationv1alpha1.Role{}).
		Complete(r)
}
