package authorizer

import (
	k "k8s.io/client-go/kubernetes"
	typedauthorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Authorizer struct {
	authClient typedauthorizationv1.AuthorizationV1Interface
}

func New() *Authorizer {
	return &Authorizer{
		authClient: k.NewForConfigOrDie(config.GetConfigOrDie()).AuthorizationV1(),
	}
}
