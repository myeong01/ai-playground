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
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

type GroupHandler struct {
	client typedauthorizationv1.AuthorizationV1alpha1Interface
	logger logger.Logger
}

func NewGroupHandler(client clientset.Interface) *GroupHandler {
	return &GroupHandler{
		client: client.AuthorizationV1alpha1(),
		logger: logger.Logger{
			Logger: logrus.New(),
		},
	}
}

func (h *GroupHandler) ListGroupInCluster(c *gin.Context) {
	groups, err := h.client.Groups("").List(c, metav1.ListOptions{})
	if err != nil {
		h.logger.Errorf("unable to list groups: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	groupList := make([]Group, len(groups.Items))
	for index, group := range groups.Items {
		groupList[index] = GroupV1alpha2Group(group)
	}
	c.JSON(http.StatusOK, ListGroupResponse{Groups: groupList})
}

func (h *GroupHandler) ListGroupInPlayground(c *gin.Context) {
	groups, err := h.client.Groups(c.Param("namespace")).List(c, metav1.ListOptions{})
	if err != nil {
		h.logger.Errorf("unable to list groups: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	groupList := make([]Group, len(groups.Items))
	for index, group := range groups.Items {
		groupList[index] = GroupV1alpha2Group(group)
	}
	c.JSON(http.StatusOK, ListGroupResponse{Groups: groupList})
}

func (h *GroupHandler) CreateGroupInPlayground(c *gin.Context) {
	var (
		namespace = c.Param("namespace")

		req  = CreateGroupRequest{}
		resp = SingleGroupResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if !IsUserSliceUnique(req.Users) {
		resp.Error = fmt.Sprintf("group create must contain unique user")
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if createdGroup, err := h.client.Groups(namespace).Create(c, &authorizationv1alpha1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
		},
		Spec: authorizationv1alpha1.GroupSpec{
			IsApproved: false,
			Users:      UserSliceToGroupV1alphaUserSlice(req.Users),
		},
	}, metav1.CreateOptions{}); err != nil {
		h.logger.Errorf("unable to create group: %v\n", err)
		resp.Error = "failed to create group"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.Group = GroupV1alpha2Group(*createdGroup)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GroupHandler) DeleteGroupInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		resp = DeleteGroupResponse{}
	)

	if err := h.client.Groups(namespace).Delete(c, objectName, metav1.DeleteOptions{}); err != nil {
		h.logger.Errorf("unable to delete group: %v\n", err)
		resp.Error = "unable to delete group"
		c.JSON(http.StatusInternalServerError, resp)
		return
	} else {
		resp.IsSuccess = true
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GroupHandler) UpdateGroupInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		req  = UpdateGroupRequest{}
		resp = SingleGroupResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	groupClient := h.client.Groups(namespace)

	if currentGroup, err := groupClient.Get(c, objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get group: %v\n", err)
		resp.Error = "unable to find group"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		if !IsUserSliceUnique(append(req.AddUser, req.RemoveUser...)) {
			resp.Error = fmt.Sprintf("group update must contain unique user")
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		for _, addUser := range req.AddUser {
			for _, user := range currentGroup.Spec.Users {
				if addUser.Name == user.Name {
					errMsg := fmt.Sprintf("group %s already has user %s", objectName, addUser.Name)
					h.logger.Errorf("unable to add user: %v\n", errMsg)
					resp.Error = errMsg
					c.JSON(http.StatusBadRequest, resp)
					return
				}
			}
		}
		currentGroup.Spec.Users = append(currentGroup.Spec.Users, UserSliceToGroupV1alphaUserSlice(req.AddUser)...)
		newUser := make([]authorizationv1alpha1.User, 0)
		for _, user := range currentGroup.Spec.Users {
			found := false
			for _, removeUser := range req.RemoveUser {
				if user.Name == removeUser.Name {
					found = true
					break
				}
			}
			if !found {
				newUser = append(newUser, user)
			}
		}
		currentGroup.Spec.Users = newUser
		if updatedGroup, err := groupClient.Update(c, currentGroup, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to update group: %v\n", err)
			resp.Error = "unable to update group"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Group = GroupV1alpha2Group(*updatedGroup)
		}
		c.JSON(http.StatusOK, resp)
		return
	}
}

func (h *GroupHandler) ApproveGroupInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		req  = ApproveGroupRequest{}
		resp = SingleGroupResponse{}
	)

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.logger.Errorf("unable to parse request: %v\n", err)
		resp.Error = "bad request"
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	groupClient := h.client.Groups(namespace)

	if currentGroup, err := groupClient.Get(c, objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get group: %v\n", err)
		resp.Error = "unable to find group"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		if req.ApproveGroup != nil {
			currentGroup.Spec.IsApproved = *req.ApproveGroup
		}
		for _, user := range req.Users {
			found := false
			for index, groupUser := range currentGroup.Spec.Users {
				if groupUser.Name == user.Name {
					currentGroup.Spec.Users[index].IsApproved = user.IsApproved
					found = true
					break
				}
			}
			if !found {
				errMsg := fmt.Sprintf("group %s doesn't contain user %s", objectName, user.Name)
				h.logger.Errorf("unable to approve user: %v\n", errMsg)
				resp.Error = errMsg
				c.JSON(http.StatusBadRequest, resp)
				return
			}
		}
		if updatedGroup, err := groupClient.Update(c, currentGroup, metav1.UpdateOptions{}); err != nil {
			h.logger.Errorf("unable to update group: %v\n", err)
			resp.Error = "unable to update group"
			c.JSON(http.StatusInternalServerError, resp)
			return
		} else {
			resp.Group = GroupV1alpha2Group(*updatedGroup)
			c.JSON(http.StatusOK, resp)
			return
		}
	}
}

func (h *GroupHandler) GetGroupInPlayground(c *gin.Context) {
	var (
		namespace  = c.Param("namespace")
		objectName = c.Param("objectName")

		resp = SingleGroupResponse{}
	)

	if group, err := h.client.Groups(namespace).Get(c, objectName, metav1.GetOptions{}); err != nil {
		h.logger.Errorf("unable to get group: %v\n", err)
		resp.Error = "unable to find group"
		if apierrs.IsNotFound(err) {
			c.JSON(http.StatusNotFound, resp)
		} else {
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	} else {
		resp.Group = GroupV1alpha2Group(*group)
	}
	c.JSON(http.StatusOK, resp)
}
