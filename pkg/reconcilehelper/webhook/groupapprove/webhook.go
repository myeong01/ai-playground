package groupapprove

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper/webhook/approve"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type GroupMutateWebhook struct {
	gvk    schema.GroupVersionKind
	client client.Client
}

func SetupGroupMutateWebhookWithManager(mgr manager.Manager) error {
	gvk, err := apiutil.GVKForObject(&authorizationv1alpha1.Group{}, mgr.GetScheme())
	if err != nil {
		return err
	}
	webhook := &GroupMutateWebhook{
		gvk:    gvk,
		client: mgr.GetClient(),
	}
	mgr.GetWebhookServer().Register(approve.GeneratePath(gvk), webhook)
	return nil
}

func (wh *GroupMutateWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	admissionReview, err := approve.GetAdmissionReviewFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	reqGroup := &authorizationv1alpha1.Group{}
	err = json.Unmarshal(admissionReview.Request.Object.Raw, &reqGroup)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	resp := admissionv1.AdmissionResponse{}
	resp.Allowed = true
	resp.UID = admissionReview.Request.UID

	patches := make([]approve.PatchOperation, 0)

	hasGroupApprovePermission, hasGroupUserApprovePermission, err := wh.checkUserHasGroupOrGroupUserApprovePermission(r.Context(), admissionReview.Request.UserInfo, reqGroup.Name, reqGroup.Namespace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to check user approve permission : %v", err)))
		return
	}

	switch admissionReview.Request.Operation {
	case admissionv1.Create:
		if !hasGroupApprovePermission {
			if reqGroup.Spec.IsApproved {
				patches = append(patches, approve.PatchOperation{
					Op:    "replace",
					Path:  "/spec/isApproved",
					Value: false,
				})
			}
			if !hasGroupUserApprovePermission {
				for index, user := range reqGroup.Spec.Users {
					if user.IsApproved {
						patches = append(patches, approve.PatchOperation{
							Op:    "replace",
							Path:  fmt.Sprintf("/spec/users/%d/isApproved", index),
							Value: false,
						})
					}
				}
			}
		}
	case admissionv1.Update:
		if !hasGroupApprovePermission {
			oldGroup := &authorizationv1alpha1.Group{}
			if err := json.Unmarshal(admissionReview.Request.OldObject.Raw, &oldGroup); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			if reqGroup.Spec.IsApproved != oldGroup.Spec.IsApproved {
				resp.Allowed = false
				resp.Result = &metav1.Status{
					Reason: metav1.StatusReasonForbidden,
				}
			}
			if !hasGroupUserApprovePermission && resp.Allowed {
				for index, user := range reqGroup.Spec.Users {
					found := false
					for _, oldGroupUser := range oldGroup.Spec.Users {
						if user.Name == oldGroupUser.Name {
							found = true
							if user.IsApproved != oldGroupUser.IsApproved {
								patches = append(patches, approve.PatchOperation{
									Op:    "replace",
									Path:  fmt.Sprintf("/spec/users/%d/isApproved", index),
									Value: oldGroupUser.IsApproved,
								})
							}
							break
						}
					}
					if !found {
						patches = append(patches, approve.PatchOperation{
							Op:    "replace",
							Path:  fmt.Sprintf("/spec/users/%d/isApproved", index),
							Value: false,
						})
					}
				}
			}
		}
	case admissionv1.Delete:
		// TODO logging
	default:
		// TODO logging
	}

	if resp.Allowed {
		resp.Result = &metav1.Status{Status: "Success"}
		if len(patches) > 0 {
			pt := admissionv1.PatchTypeJSONPatch
			resp.PatchType = &pt
			resp.Patch, err = json.Marshal(patches)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to marshal patches : %v", err)))
				return
			}
		}
	}
	admissionReview.Response = &resp
	marshaledAdmReview, err := json.Marshal(admissionReview)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to marshal admissionReview : %v", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(marshaledAdmReview)
	return
}

func (wh *GroupMutateWebhook) checkUserHasGroupOrGroupUserApprovePermission(ctx context.Context, userInfo authenticationv1.UserInfo, name, namespace string) (bool, bool, error) {
	isGroupApproveAllowed, err := wh.checkUserHasGroupApprovePermission(ctx, userInfo, name, namespace)
	if err != nil {
		return false, false, err
	}
	if isGroupApproveAllowed {
		return true, true, nil
	}
	isGroupUserApproveAllowed, err := wh.checkUserHasGroupUserApprovePermission(ctx, userInfo, name, namespace)
	return false, isGroupUserApproveAllowed, err
}

func (wh *GroupMutateWebhook) checkUserHasGroupApprovePermission(ctx context.Context, userInfo authenticationv1.UserInfo, name, namespace string) (bool, error) {
	subjectAccessReview := &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Name:        name,
				Namespace:   namespace,
				Verb:        "update",
				Group:       wh.gvk.Group,
				Version:     wh.gvk.Version,
				Resource:    "groups",
				Subresource: "approval",
			},
			User:   userInfo.Username,
			Groups: userInfo.Groups,
			UID:    userInfo.UID,
		},
	}
	err := wh.client.Create(ctx, subjectAccessReview)
	if err != nil {
		return false, err
	}
	return subjectAccessReview.Status.Allowed, nil
}

func (wh *GroupMutateWebhook) checkUserHasGroupUserApprovePermission(ctx context.Context, userInfo authenticationv1.UserInfo, name, namespace string) (bool, error) {
	group := &authorizationv1alpha1.Group{}
	if err := wh.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, group); err != nil {
		return false, err
	}
	for _, user := range group.Status.ApprovedUsers {
		if userInfo.Username == user.Name {
			var rules []rbacv1.PolicyRule
			switch user.Role.Type {
			case "ClusterRole":
				clusterRole := &authorizationv1alpha1.ClusterRole{}
				if err := wh.client.Get(ctx, types.NamespacedName{Name: user.Role.Name}, clusterRole); err != nil {
					return false, err
				}
				rules = clusterRole.Status.Rules
			case "Role":
				role := &authorizationv1alpha1.Role{}
				if err := wh.client.Get(ctx, types.NamespacedName{Name: user.Role.Name, Namespace: namespace}, role); err != nil {
					return false, err
				}
				rules = role.Status.Rules
			default:
				return false, errors.New("unknown role type")
			}
			for _, rule := range rules {
				if len(rule.APIGroups) > 0 && rule.APIGroups[0] == wh.gvk.Group && len(rule.Resources) > 0 && rule.Resources[0] == "groups/approval" && len(rule.Verbs) > 0 && rule.Verbs[0] == "update" {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, nil
}
