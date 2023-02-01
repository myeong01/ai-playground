package authorizer

import (
	"context"
	"errors"
	playground "github.com/myeong01/ai-playground/pkg/clientset/clientset"
	typedauthorizationv1alpha1 "github.com/myeong01/ai-playground/pkg/clientset/clientset/typed/authorization/v1alpha1"
	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k "k8s.io/client-go/kubernetes"
	typedauthorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Authorizer struct {
	kubeAuthClient         typedauthorizationv1.AuthorizationV1Interface
	aiPlaygroundAuthClient typedauthorizationv1alpha1.AuthorizationV1alpha1Interface
}

func New() *Authorizer {
	conf := config.GetConfigOrDie()
	kubeClient := k.NewForConfigOrDie(conf)
	aiPlaygroundClient := playground.NewForConfigOrDie(conf)
	return &Authorizer{
		kubeAuthClient:         kubeClient.AuthorizationV1(),
		aiPlaygroundAuthClient: aiPlaygroundClient.AuthorizationV1alpha1(),
	}
}

func (a *Authorizer) Validate(ctx context.Context, checker *Checker) (bool, error) {
	isAllowedAtKubeWide, err := a.validateWithKubeWide(ctx, checker)
	if err != nil {
		return false, err
	}
	if isAllowedAtKubeWide {
		return true, nil
	}
	if checker.GroupName != "" {
		return a.validateWithPlaygroundWide(ctx, checker)
	}
	return false, nil
}

func (a *Authorizer) validateWithKubeWide(ctx context.Context, checker *Checker) (bool, error) {
	resp, err := a.kubeAuthClient.SubjectAccessReviews().Create(ctx, &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			ResourceAttributes: checker.ResourceAttributes,
			User:               checker.UserInfo.Username,
			Groups:             checker.UserInfo.Groups,
			UID:                checker.UserInfo.UID,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return resp.Status.Allowed, err
}

func (a *Authorizer) validateWithPlaygroundWide(ctx context.Context, checker *Checker) (bool, error) {
	group, err := a.aiPlaygroundAuthClient.Groups(checker.ResourceAttributes.Namespace).Get(ctx, checker.ResourceAttributes.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	for _, user := range group.Status.ApprovedUsers {
		if user.Name == checker.UserInfo.Username {

			var rules []rbacv1.PolicyRule
			switch user.Role.Type {
			case "ClusterRole":
				clusterRole, err := a.aiPlaygroundAuthClient.ClusterRoles().Get(ctx, user.Role.Name, metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				rules = clusterRole.Status.Rules
			case "Role":
				role, err := a.aiPlaygroundAuthClient.Roles(checker.ResourceAttributes.Namespace).Get(ctx, checker.ResourceAttributes.Name, metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				rules = role.Status.Rules
			default:
				return false, errors.New("unknown role type")
			}
			for _, rule := range rules {
				if len(rule.APIGroups) > 0 && rule.APIGroups[0] == checker.ResourceAttributes.Group && len(rule.Resources) > 0 && rule.Resources[0] == checker.ResourceAttributes.Resource && len(rule.Verbs) > 0 && rule.Verbs[0] == checker.ResourceAttributes.Verb {
					return true, nil
				}

			}
		}
	}
	return false, nil
}
