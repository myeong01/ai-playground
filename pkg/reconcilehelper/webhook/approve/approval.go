package approve

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type ApprovalObject interface {
	runtime.Object
	IsApproved() bool
	ApprovalPath() string
}
