package approve

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
)

type Builder struct {
	mgr            manager.Manager
	approvalObject ApprovalObject
}

func NewWebhookManagedBy(mgr manager.Manager) *Builder {
	return &Builder{
		mgr: mgr,
	}
}

func (b *Builder) For(approvalObject ApprovalObject) *Builder {
	b.approvalObject = approvalObject
	return b
}

func (b *Builder) Complete() error {
	webhook, err := NewWebhook(b.approvalObject, b.mgr)
	gvk, err := apiutil.GVKForObject(b.approvalObject, b.mgr.GetScheme())
	if err != nil {
		return err
	}

	b.mgr.GetWebhookServer().Register(generatePath(gvk), webhook)
	return nil
}

func generatePath(gvk schema.GroupVersionKind) string {
	return "/playground-mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" + gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
