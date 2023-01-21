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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var containersnapshotlog = logf.Log.WithName("containersnapshot-resource")

func (r *ContainerSnapshot) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-container-ai-playground-io-v1alpha1-containersnapshot,mutating=false,failurePolicy=fail,sideEffects=None,groups=container.ai-playground.io,resources=containersnapshots,verbs=create;update,versions=v1alpha1,name=vcontainersnapshot.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ContainerSnapshot{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ContainerSnapshot) ValidateCreate() error {
	containersnapshotlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ContainerSnapshot) ValidateUpdate(old runtime.Object) error {
	containersnapshotlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ContainerSnapshot) ValidateDelete() error {
	containersnapshotlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
