package approve

import (
	"context"
	"encoding/json"
	"fmt"
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

type Webhook[T ApprovalObject] struct {
	gvk        schema.GroupVersionKind
	authClient typedauthorizationv1.AuthorizationV1Interface
}

func NewWebhook[T ApprovalObject](object T, mgr manager.Manager) (*Webhook[T], error) {
	gvk, err := apiutil.GVKForObject(object, mgr.GetScheme())
	if err != nil {
		return nil, err
	}
	return &Webhook[T]{
		gvk:        gvk,
		authClient: k.NewForConfigOrDie(config.GetConfigOrDie()).AuthorizationV1(),
	}, nil
}

func (wh *Webhook[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	admissionReview, err := GetAdmissionReviewFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var reqObject T
	err = json.Unmarshal(admissionReview.Request.Object.Raw, &reqObject)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("bad request : %v", err)))
		return
	}

	resp := admissionv1.AdmissionResponse{}
	resp.Allowed = true
	resp.UID = admissionReview.Request.UID
	pT := admissionv1.PatchTypeJSONPatch
	resp.PatchType = &pT
	resp.Result = &metav1.Status{Status: "Success"}

	ctx := r.Context()
	reqName, reqNamespace := admissionReview.Request.Name, admissionReview.Request.Namespace
	switch admissionReview.Request.Operation {
	case admissionv1.Create:
		isApproved := reqObject.IsApproved()
		if isApproved {
			err = wh.fillRespWithPermissionCheck(ctx, reqName, reqNamespace, admissionReview.Request.UserInfo, resp, reqObject.ApprovalPath(), false)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to valid permission : %v", err)))
			}
		}
	case admissionv1.Update:
		var oldObject T
		err = json.Unmarshal(admissionReview.Request.OldObject.Raw, &oldObject)
		isApprovedNow := reqObject.IsApproved()
		isApproveBefore := oldObject.IsApproved()
		if isApprovedNow != isApproveBefore {
			// TODO isApproveBefore field 를 유지하는 것이 맞는지 고민 필요
			err = wh.fillRespWithPermissionCheck(ctx, reqName, reqNamespace, admissionReview.Request.UserInfo, resp, reqObject.ApprovalPath(), isApproveBefore)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to valid permission : %v", err)))
			}
		}
	case admissionv1.Delete:
		// TODO logging
	default:
		// TODO logging
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

func (wh *Webhook[T]) fillRespWithPermissionCheck(ctx context.Context, name, namespace string, userInfo authenticationv1.UserInfo, resp admissionv1.AdmissionResponse, patchPath string, defaultApprovalValue bool) error {
	valid, err := wh.validation(ctx, name, namespace, userInfo)
	if err != nil {
		return err
	}
	if valid {
		return nil
	} else {
		resp.Result = &metav1.Status{Status: "Approve field replaced", Reason: "permission denied"}
		patches := []PatchOperation{
			{
				Op:    "replace",
				Path:  patchPath,
				Value: defaultApprovalValue,
			},
		}
		resp.Patch, err = json.Marshal(patches)
		return err
	}
}

func (wh *Webhook[T]) validation(ctx context.Context, name, namespace string, userInfo authenticationv1.UserInfo) (bool, error) {
	resp, err := wh.authClient.SubjectAccessReviews().Create(ctx, &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			ResourceAttributes: wh.generateResourceAttributes(name, namespace),
			User:               userInfo.Username,
			Groups:             userInfo.Groups,
			UID:                userInfo.UID,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return resp.Status.Allowed, nil
}

func (wh *Webhook[T]) generateResourceAttributes(name, namespace string) *authorizationv1.ResourceAttributes {
	return &authorizationv1.ResourceAttributes{
		Name:        name,
		Namespace:   namespace,
		Verb:        "update",
		Group:       wh.gvk.Group,
		Version:     wh.gvk.Version,
		Resource:    wh.gvk.Kind,
		Subresource: "approval",
	}
}
