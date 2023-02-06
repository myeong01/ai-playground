package v1alpha1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/clientset/clientset"
	typedauthorizationv1 "github.com/myeong01/ai-playground/pkg/clientset/clientset/typed/authorization/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/logger"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

type RoleHandler struct {
	client typedauthorizationv1.AuthorizationV1alpha1Interface
	logger logger.Logger
}

func NewRoleHandler(client clientset.Interface) *RoleHandler {
	return &RoleHandler{
		client: client.AuthorizationV1alpha1(),
		logger: logger.Logger{
			Logger: logrus.New(),
		},
	}
}

func (h *RoleHandler) ListRoleInCluster(c *gin.Context) {
	var (
		resp = ListRoleResponse{}
	)

	roles, err := h.client.Roles("").List(c, metav1.ListOptions{})
	if err != nil {
		h.logger.Errorf("unable to list roles: %v\n", err)
		resp.Error = "unable to list roles"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	roleList := make([]Role, len(roles.Items))
	for index, role := range roles.Items {
		roleList[index] = RoleV1Alpha2Role(role)
	}
	resp.Roles = roleList
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) ListRoleInPlayground(c *gin.Context) {
	var (
		resp = ListRoleResponse{}
	)

	roles, err := h.client.Roles(c.Param("namespace")).List(c, metav1.ListOptions{})
	if err != nil {
		h.logger.Errorf("unable to list roles: %v\n", err)
		resp.Error = "unable to list roles"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	roleList := make([]Role, len(roles.Items))
	for index, role := range roles.Items {
		roleList[index] = RoleV1Alpha2Role(role)
	}
	resp.Roles = roleList
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) ListRoleInGroup(c *gin.Context) {
	var (
		resp = ListRoleResponse{}
	)

	roles, err := h.client.Roles(c.Param("namespace")).List(c, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("group=%s", c.Param("group")),
	})
	if err != nil {
		h.logger.Errorf("unable to list roles: %v\n", err)
		resp.Error = "unable to list roles"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	roleList := make([]Role, len(roles.Items))
	for index, role := range roles.Items {
		roleList[index] = RoleV1Alpha2Role(role)
	}
	resp.Roles = roleList
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) GetRoleInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		resp = SingleRoleResponse{}
	)

	if role, err := h.client.Roles(namespace).Get(c, getRoleNameInPlayground(objectName), metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get role: %v\n", err)
		resp.Error = "unable to find role"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		resp.Role = RoleV1Alpha2Role(*role)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) GetRoleInGroup(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")
		groupName  = c.Param("group")

		resp = SingleRoleResponse{}
	)

	if role, err := h.client.Roles(namespace).Get(c, getRoleNameInGroup(objectName, groupName), metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get role: %v\n", err)
		resp.Error = "unable to find role"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		resp.Role = RoleV1Alpha2Role(*role)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) CreateRoleInPlayground(c *gin.Context) {
	var (
		namespace = c.Param("namespace")

		req  = CreateRoleRequest{}
		resp = SingleRoleResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if createdRole, err := h.client.Roles(namespace).Create(c, &authorizationv1alpha1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getRoleNameInPlayground(req.Role.Name),
			Namespace: namespace,
		},
		Spec: authorizationv1alpha1.RoleSpec{
			IsApproved: false,
			ParentRole: &req.Role.ParentRole,
			Rules:      req.Role.Rules,
		},
	}, metav1.CreateOptions{}); err != nil {
		h.logger.Errorf("unable to create role: %v\n", err)
		resp.Error = "failed to create role"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.Role = RoleV1Alpha2Role(*createdRole)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) CreateRoleInGroup(c *gin.Context) {
	var (
		namespace = c.Param("namespace")
		groupName = c.Param("group")

		req  = CreateRoleRequest{}
		resp = SingleRoleResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if createdRole, err := h.client.Roles(namespace).Create(c, &authorizationv1alpha1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getRoleNameInGroup(req.Role.Name, groupName),
			Namespace: namespace,
			Labels: map[string]string{
				"group": groupName,
			},
		},
		Spec: authorizationv1alpha1.RoleSpec{
			IsApproved: false,
			ParentRole: &req.Role.ParentRole,
			Rules:      req.Role.Rules,
		},
	}, metav1.CreateOptions{}); err != nil {
		h.logger.Errorf("unable to create role: %v\n", err)
		resp.Error = "failed to create role"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.Role = RoleV1Alpha2Role(*createdRole)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) UpdateRoleInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		req  = UpdateRoleRequest{}
		resp = SingleRoleResponse{}

		roleClient = h.client.Roles(namespace)
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if currentRole, err := roleClient.Get(c, getRoleNameInPlayground(objectName), metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get role: %v\n", err)
		resp.Error = "unable to find role"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		currentRole.Spec.Rules = append(currentRole.Spec.Rules, req.AddRule...)
		removeRuleMarshaledMap := make(map[string]struct{})
		for _, rule := range req.RemoveRule {
			if marshaledRule, err := json.Marshal(rule); err != nil {
				h.logger.Errorf("unable to marshal rule: %v\n", err)
				resp.Error = "failed to update rule on removing existing rule"
				c.JSON(http.StatusInternalServerError, resp)
				return
			} else {
				removeRuleMarshaledMap[string(marshaledRule)] = struct{}{}
			}
		}
		newRule := make([]rbacv1.PolicyRule, 0)
		for _, rule := range currentRole.Spec.Rules {
			if marshaledRule, err := json.Marshal(rule); err != nil {
				h.logger.Errorf("unable to marshal rule: %v\n", err)
				resp.Error = "failed to update rule on removing existing rule"
				c.JSON(http.StatusInternalServerError, resp)
				return
			} else if _, exist := removeRuleMarshaledMap[string(marshaledRule)]; !exist {
				newRule = append(newRule, rule)
			}
		}
		if !IsRuleSliceUnique(newRule) {
			h.logger.Errorf("rule is not unique after role update")
			resp.Error = "role cannot own duplicated rule"
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		currentRole.Spec.Rules = newRule
		if updatedRole, err := roleClient.Update(c, currentRole, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to update role: %v\n", err)
			resp.Error = "unable to update role"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Role = RoleV1Alpha2Role(*updatedRole)
		}
		c.JSON(http.StatusOK, resp)
		return
	}
}

func (h *RoleHandler) UpdateRoleInGroup(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		groupName  = c.Param("group")
		objectName = c.Param("objectName")

		req  = UpdateRoleRequest{}
		resp = SingleRoleResponse{}

		roleClient = h.client.Roles(namespace)
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if currentRole, err := roleClient.Get(c, getRoleNameInGroup(objectName, groupName), metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get role: %v\n", err)
		resp.Error = "unable to find role"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		currentRole.Spec.Rules = append(currentRole.Spec.Rules, req.AddRule...)
		removeRuleMarshaledMap := make(map[string]struct{})
		for _, rule := range req.RemoveRule {
			if marshaledRule, err := json.Marshal(rule); err != nil {
				h.logger.Errorf("unable to marshal rule: %v\n", err)
				resp.Error = "failed to update rule on removing existing rule"
				c.JSON(http.StatusInternalServerError, resp)
				return
			} else {
				removeRuleMarshaledMap[string(marshaledRule)] = struct{}{}
			}
		}
		newRule := make([]rbacv1.PolicyRule, 0)
		for _, rule := range currentRole.Spec.Rules {
			if marshaledRule, err := json.Marshal(rule); err != nil {
				h.logger.Errorf("unable to marshal rule: %v\n", err)
				resp.Error = "failed to update rule on removing existing rule"
				c.JSON(http.StatusInternalServerError, resp)
				return
			} else if _, exist := removeRuleMarshaledMap[string(marshaledRule)]; !exist {
				newRule = append(newRule, rule)
			}
		}
		if !IsRuleSliceUnique(newRule) {
			h.logger.Errorf("rule is not unique after role update")
			resp.Error = "role cannot own duplicated rule"
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		currentRole.Spec.Rules = newRule
		if updatedRole, err := roleClient.Update(c, currentRole, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to update role: %v\n", err)
			resp.Error = "unable to update role"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Role = RoleV1Alpha2Role(*updatedRole)
		}
		c.JSON(http.StatusOK, resp)
		return
	}
}

func (h *RoleHandler) DeleteRoleInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		resp = DeleteRoleResponse{}
	)

	if err := h.client.Roles(namespace).Delete(c, getRoleNameInPlayground(objectName), metav1.DeleteOptions{}); err != nil && !apierrs.IsNotFound(err) {
		h.logger.Errorf("unable to delete role: %v\n", err)
		resp.Error = "unable to delete role"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.IsSuccess = true
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) DeleteRoleInGroup(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		groupName  = c.Param("group")
		objectName = c.Param("objectName")

		resp = DeleteRoleResponse{}
	)

	if err := h.client.Roles(namespace).Delete(c, getRoleNameInGroup(objectName, groupName), metav1.DeleteOptions{}); err != nil && !apierrs.IsNotFound(err) {
		h.logger.Errorf("unable to delete role: %v\n", err)
		resp.Error = "unable to delete role"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.IsSuccess = true
	}
	c.JSON(http.StatusOK, resp)
}

func (h *RoleHandler) ApproveRoleInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		req  = ApproveRoleRequest{}
		resp = SingleRoleResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	roleClient := h.client.Roles(namespace)

	if currentRole, err := roleClient.Get(c, getRoleNameInPlayground(objectName), metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get role: %v\n", err)
		resp.Error = "unable to find role"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		currentRole.Spec.IsApproved = req.IsApproved
		if updatedRole, err := roleClient.Update(c, currentRole, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to update role: %v\n", err)
			resp.Error = "unable to update role"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Role = RoleV1Alpha2Role(*updatedRole)
			c.JSON(http.StatusOK, resp)
			return
		}
	}
}

func (h *RoleHandler) ApproveRoleInGroup(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		groupName  = c.Param("group")
		objectName = c.Param("objectName")

		req  = ApproveRoleRequest{}
		resp = SingleRoleResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	roleClient := h.client.Roles(namespace)

	if currentRole, err := roleClient.Get(c, getRoleNameInGroup(objectName, groupName), metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get role: %v\n", err)
		resp.Error = "unable to find role"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		currentRole.Spec.IsApproved = req.IsApproved
		if updatedRole, err := roleClient.Update(c, currentRole, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to update role: %v\n", err)
			resp.Error = "unable to update role"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Role = RoleV1Alpha2Role(*updatedRole)
			c.JSON(http.StatusOK, resp)
			return
		}
	}
}
