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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
)

// GroupReconciler reconciles a Group object
type GroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=groups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=groups/approval,verbs=update
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=groups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=authorization.ai-playground.io,resources=groups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Group object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	group := &authorizationv1alpha1.Group{}
	if err := r.Get(ctx, req.NamespacedName, group); err != nil {
		return ctrl.Result{}, reconcilehelper.GetErrorExceptNotFound(err, "unable to fetch Group", logger)
	}

	if !group.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if !group.Spec.IsApproved {
		return ctrl.Result{}, nil
	}

	if updatedApprovedUser, isUpdated := checkUserUpdate(group.Spec.Users, group.Status.ApprovedUsers); isUpdated {
		group.Status.ApprovedUsers = updatedApprovedUser
		if err := r.Status().Update(ctx, group); err != nil {
			logger.Error(err, "unable to update Group status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func checkUserUpdate(users []authorizationv1alpha1.User, approvedUsers []authorizationv1alpha1.ApprovedUser) ([]authorizationv1alpha1.ApprovedUser, bool) {
	updatedApprovedUsers := make([]authorizationv1alpha1.ApprovedUser, len(users))
	approvedUserLen := len(approvedUsers)
	updatedApprovedUsersIndex := 0
	approvedUsersIndex := 0
	needUpdate := false
	for _, user := range users {
		if approvedUsersIndex < approvedUserLen {
			if user.Name != approvedUsers[approvedUsersIndex].Name {
				if user.IsApproved {
					updatedApprovedUsers[updatedApprovedUsersIndex] = userToApprovedUser(user)
					updatedApprovedUsersIndex++
					needUpdate = true
				}
			} else {
				if user.IsApproved {
					updatedApprovedUsers[updatedApprovedUsersIndex] = userToApprovedUser(user)
					updatedApprovedUsersIndex++
				} else {
					needUpdate = true
				}
				approvedUsersIndex++
			}
		} else {
			if user.IsApproved {
				updatedApprovedUsers[updatedApprovedUsersIndex] = userToApprovedUser(user)
				updatedApprovedUsersIndex++
				needUpdate = true
			}
		}
	}
	return updatedApprovedUsers[:updatedApprovedUsersIndex], needUpdate
}

func userToApprovedUser(user authorizationv1alpha1.User) authorizationv1alpha1.ApprovedUser {
	return authorizationv1alpha1.ApprovedUser{
		Name: user.Name,
		Role: user.Role,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authorizationv1alpha1.Group{}).
		Complete(r)
}
