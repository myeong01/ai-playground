package approve

import (
	"encoding/json"
	"errors"
	"fmt"
	admissionv1 "k8s.io/api/admission/v1"
	"net/http"
)

func GetAdmissionReviewFromRequest(request *http.Request) (*admissionv1.AdmissionReview, error) {
	if err := validateRequest(request); err != nil {
		return nil, err
	}

	admReview := admissionv1.AdmissionReview{}

	err := json.NewDecoder(request.Body).Decode(&admReview)
	if err != nil {
		return nil, err
	}

	return &admReview, nil
}

func validateRequest(req *http.Request) error {
	if req.Method != http.MethodPost {
		return fmt.Errorf("expect http method 'POST' but got '%s'", req.Method)
	}
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("expected http Content-Type 'application/json', got '%s'", contentType)
	}
	if req.Body == nil {
		return errors.New("empty body")
	}
	return nil
}
