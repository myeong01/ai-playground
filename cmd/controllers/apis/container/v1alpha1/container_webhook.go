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
	"github.com/myeong01/ai-playground/pkg/reconcilehelper/webhook/approve"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var containerlog = logf.Log.WithName("container-resource")

func (r *Container) SetupWebhookWithManager(mgr ctrl.Manager) error {
	err := approve.NewWebhookManagedBy(mgr, r, "containers").
		Complete()
	if err != nil {
		return err
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-container-ai-playground-io-v1alpha1-container,mutating=false,failurePolicy=fail,sideEffects=None,groups=container.ai-playground.io,resources=containers,verbs=create;update,versions=v1alpha1,name=vcontainer.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/playground-mutate-container-ai-playground-io-v1alpha1-container,mutating=true,failurePolicy=fail,sideEffects=None,groups=container.ai-playground.io,resources=containers,verbs=create;update,versions=v1alpha1,name=mcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Container{}
var _ approve.ApprovalObject = &Container{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Container) ValidateCreate() error {
	containerlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Container) ValidateUpdate(old runtime.Object) error {
	containerlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Container) ValidateDelete() error {
	containerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Container) IsApproved() bool {
	return r.Spec.IsApproved
}

func (r *Container) ApprovalPath() string {
	return "/spec/isApproved"
}
