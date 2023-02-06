package v1alpha1

import (
	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
	"strings"
)

type ListGroupResponse struct {
	Groups []Group `json:"groups,omitempty"`
	Error  string  `json:"error,omitempty"`
}

type Group struct {
	Status     string                       `json:"status"`
	Name       string                       `json:"name"`
	Playground string                       `json:"playground"`
	Users      []authorizationv1alpha1.User `json:"users,omitempty"`
}

func GroupV1alpha2Group(groupV1alpha1 authorizationv1alpha1.Group) Group {
	group := Group{
		Name:       groupV1alpha1.Name,
		Playground: strings.TrimPrefix(groupV1alpha1.Namespace, "pg-"),
		Users:      groupV1alpha1.Spec.Users,
	}
	if groupV1alpha1.Spec.IsApproved {
		group.Status = "working"
	} else {
		group.Status = "ready for approve"
	}
	return group
}

type User struct {
	Name string                             `json:"name"`
	Role authorizationv1alpha1.RoleSelector `json:"role"`
}

type CreateGroupRequest struct {
	Name  string `json:"name"`
	Users []User `json:"users,omitempty"`
}

type SingleGroupResponse struct {
	Group Group  `json:"group,omitempty"`
	Error string `json:"error,omitempty"`
}

type DeleteGroupResponse struct {
	IsSuccess bool   `json:"isSuccess"`
	Error     string `json:"error,omitempty"`
}

type UpdateGroupRequest struct {
	AddUser    []User `json:"addUser,omitempty"`
	RemoveUser []User `json:"removeUser,omitempty"`
}

type ApproveUser struct {
	Name       string `json:"name"`
	IsApproved bool   `json:"isApproved"`
}

type ApproveGroupRequest struct {
	Users        []ApproveUser `json:"users,omitempty"`
	ApproveGroup *bool         `json:"approveGroup,omitempty"`
}

func User2GroupV1alpha(user User) authorizationv1alpha1.User {
	return authorizationv1alpha1.User{
		Name:       user.Name,
		Role:       user.Role,
		IsApproved: false,
	}
}

func UserSliceToGroupV1alphaUserSlice(users []User) []authorizationv1alpha1.User {
	authUsers := make([]authorizationv1alpha1.User, len(users))
	for index, user := range users {
		authUsers[index] = User2GroupV1alpha(user)
	}
	return authUsers
}

func IsUserSliceUnique(users []User) bool {
	usersMap := make(map[string]struct{})

	for _, user := range users {
		if _, exist := usersMap[user.Name]; exist {
			return false
		} else {
			usersMap[user.Name] = struct{}{}
		}
	}
	return true
}
