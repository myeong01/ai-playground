package v1alpha1

import (
	"encoding/json"
	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type ListRoleResponse struct {
	Roles []Role `json:"roles,omitempty"`
	Error string `json:"error,omitempty"`
}

type SingleRoleResponse struct {
	Role  Role   `json:"role,omitempty"`
	Error string `json:"error,omitempty"`
}

type CreateRoleRequest struct {
	Role Role `json:"role"`
}

type DeleteRoleResponse struct {
	IsSuccess bool   `json:"isSuccess"`
	Error     string `json:"error,omitempty"`
}

type UpdateRoleRequest struct {
	AddRule    []rbacv1.PolicyRule `json:"addRule,omitempty"`
	RemoveRule []rbacv1.PolicyRule `json:"removeRule,omitempty"`
}

type ApproveRoleRequest struct {
	IsApproved bool `json:"isApproved"`
}

type Role struct {
	Name       string                             `json:"name"`
	ParentRole authorizationv1alpha1.RoleSelector `json:"parentRole"`
	Rules      []rbacv1.PolicyRule                `json:"rules,omitempty"`
	GroupName  string                             `json:"group,omitempty"`
}

// playground-<name>
// group-<group name>-<name>

func RoleV1Alpha2Role(roleV1Alpha authorizationv1alpha1.Role) Role {
	role := Role{
		Rules: roleV1Alpha.Spec.Rules,
	}
	if roleV1Alpha.Spec.ParentRole != nil {
		role.ParentRole = *roleV1Alpha.Spec.ParentRole
	}
	role.Name = getClearRoleName(roleV1Alpha)
	role.GroupName, _ = getGroupName(roleV1Alpha.ObjectMeta)
	return role
}

func getClearRoleName(roleV1Alpha authorizationv1alpha1.Role) string {
	if groupName, exist := getGroupName(roleV1Alpha.ObjectMeta); exist {
		return strings.TrimPrefix(roleV1Alpha.Name, "group-"+groupName+"-")
	} else {
		return strings.TrimPrefix(roleV1Alpha.Name, "playground-")
	}
}

func getRoleNameInPlayground(name string) string {
	return "playground-" + name
}

func getRoleNameInGroup(name, groupName string) string {
	return "group-" + groupName + "-" + name
}

func getGroupName(meta metav1.ObjectMeta) (string, bool) {
	groupName, exist := meta.Labels["group"]
	return groupName, exist
}

func IsRuleSliceUnique(rules []rbacv1.PolicyRule) bool {
	rulesMap := make(map[string]struct{})

	for _, rule := range rules {
		if marshaledRule, err := json.Marshal(rule); err != nil {
			return false
		} else if _, exist := rulesMap[string(marshaledRule)]; exist {
			return false
		} else {
			rulesMap[string(marshaledRule)] = struct{}{}
		}
	}
	return true
}
