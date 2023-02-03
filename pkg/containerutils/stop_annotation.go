package containerutils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	StopAnnotation         = "ai-playground-resource-stopped"
	LastActivityAnnotation = "container.ai-playground.ai/last_activity"
)

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
