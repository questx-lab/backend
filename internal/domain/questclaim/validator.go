package questclaim

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/questx-lab/backend/pkg/router"
)

// VisitLink Validator
type visitLinkValidator struct {
	Link string `json:"link,omitempty"`
}

func newVisitLinkValidator(ctx router.Context, data string) (*visitLinkValidator, error) {
	visitLink := visitLinkValidator{}
	err := json.Unmarshal([]byte(data), &visitLink)
	if err != nil {
		return nil, err
	}

	if visitLink.Link == "" {
		return nil, errors.New("Not found link in validation data")
	}

	_, err = url.ParseRequestURI(visitLink.Link)
	if err != nil {
		return nil, err
	}

	return &visitLink, nil
}

func (v *visitLinkValidator) Validate(router.Context, string) (ValidateResult, error) {
	return Accepted, nil
}

// Text Validator
// TODO: Add retry_after when the claimed quest is rejected by auto validate.
type textValidator struct {
	AutoValidate bool   `json:"auto_validate"`
	Answer       string `json:"answer"`
}

func newTextValidator(ctx router.Context, data string) (*textValidator, error) {
	text := textValidator{}
	err := json.Unmarshal([]byte(data), &text)
	if err != nil {
		return nil, err
	}

	return &text, nil
}

func (v *textValidator) Validate(ctx router.Context, input string) (ValidateResult, error) {
	if !v.AutoValidate {
		return NeedManualReview, nil
	}

	if v.Answer != input {
		return Rejected, nil
	}

	return Accepted, nil
}
