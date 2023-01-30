package groupapprove

import (
	"context"
	"encoding/json"
	"fmt"
	authorizationv1alpha1 "github.com/myeong01/ai-playground/cmd/controllers/apis/authorization/v1alpha1"
	"github.com/myeong01/ai-playground/pkg/reconcilehelper/webhook/approve"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k "k8s.io/client-go/kubernetes"
	typedauthorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type GroupMutateWebhook struct {
	gvk        schema.GroupVersionKind
	authClient typedauthorizationv1.AuthorizationV1Interface
}

func SetupGroupMutateWebhookWithManager(mgr manager.Manager) error {
	gvk, err := apiutil.GVKForObject(&authorizationv1alpha1.Group{}, mgr.GetScheme())
	if err != nil {
		return err
	}
	webhook := &GroupMutateWebhook{
		gvk:        gvk,
		authClient: k.NewForConfigOrDie(config.GetConfigOrDie()).AuthorizationV1(),
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
	return false, false, nil
}

func (wh *GroupMutateWebhook) checkUserHasGroupApprovePermission(ctx context.Context, userInfo authenticationv1.UserInfo, name, namespace string) (bool, error) {
	resp, err := wh.authClient.SubjectAccessReviews().Create(ctx, &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Name:        name,
				Namespace:   namespace,
				Verb:        "update",
				Group:       wh.gvk.Group,
				Version:     wh.gvk.Version,
				Resource:    wh.gvk.Kind,
				Subresource: "approval",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return resp.Status.Allowed, nil
}

func (wh *GroupMutateWebhook) checkUserHasGroupUserApprovePermission(ctx context.Context, userInfo authenticationv1.UserInfo, name, namespace string) (bool, error) {
	return false, nil
}
