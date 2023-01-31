package approve

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
)

type Builder[T ApprovalObject] struct {
	mgr            manager.Manager
	approvalObject T
	resourceName   string
}

func NewWebhookManagedBy[T ApprovalObject](mgr manager.Manager, approvalObject T, resourceName string) *Builder[T] {
	return &Builder[T]{
		mgr:            mgr,
		approvalObject: approvalObject,
		resourceName:   resourceName,
	}
}

//func (b *Builder[T]) For(approvalObject T) *Builder[T] {
//	b.approvalObject = approvalObject
//	return b
//}

func (b *Builder[T]) Complete() error {
	webhook, err := NewWebhook(b.approvalObject, b.mgr, b.resourceName)
	gvk, err := apiutil.GVKForObject(b.approvalObject, b.mgr.GetScheme())
	if err != nil {
		return err
	}

	b.mgr.GetWebhookServer().Register(GeneratePath(gvk), webhook)
	return nil
}

func GeneratePath(gvk schema.GroupVersionKind) string {
	return "/playground-mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" + gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
