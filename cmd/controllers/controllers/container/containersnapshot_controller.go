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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	containerv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/container/v1alpha1"
)

// ContainerSnapshotReconciler reconciles a ContainerSnapshot object
type ContainerSnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=container.ai-playground.io,resources=containersnapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=container.ai-playground.io,resources=containersnapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=container.ai-playground.io,resources=containersnapshots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ContainerSnapshot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ContainerSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	snapshot := &containerv1alpha1.ContainerSnapshot{}
	if err := r.Get(ctx, req.NamespacedName, snapshot); err != nil {
		logger.Error(err, "unable to fetch container snapshot for container snapshot")
		return ctrl.Result{}, err
	}

	container := &containerv1alpha1.Container{}
	if err := r.Get(ctx, types.NamespacedName{Name: snapshot.Spec.ContainerName, Namespace: snapshot.Namespace}, container); err != nil {
		logger.Error(err, "unable to fetch container for container snapshot")
		return ctrl.Result{}, err
	}

	if err := ctrl.SetControllerReference(container, snapshot, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference to Snapshot")
		return ctrl.Result{}, err
	}

	versionTargetStatusList := snapshot.Spec.VersionedNames
	versionCurrentStatusList := snapshot.Status.Snapshots

	targetStatusLen := len(versionTargetStatusList)
	targetStatusIndex := 0
	currentStatusLen := len(versionCurrentStatusList)
	currentStatusIndex := 0

	for targetStatusIndex < targetStatusLen && currentStatusIndex < currentStatusLen {
		if versionTargetStatusList[targetStatusIndex] == versionCurrentStatusList[currentStatusIndex].Name {
			if versionCurrentStatusList[currentStatusIndex].CommitId == "" && !versionCurrentStatusList[currentStatusIndex].Failed {
				// TODO logic for create new version request to snapshot worker daemon then parse request and handle result
			}
			targetStatusIndex++
		} else {
			// TODO logic for delete existing version request to snapshot worker daemon then parse request and handle result
		}
		currentStatusIndex++
	}

	for targetStatusIndex < targetStatusLen {
		// TODO logic for delete existing version request to snapshot worker daemon then parse request and handle result
	}

	for currentStatusIndex < currentStatusLen {
		// TODO logic for create new version request to snapshot worker daemon then parse request and handle result
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContainerSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&containerv1alpha1.ContainerSnapshot{}).
		Complete(r)
}
