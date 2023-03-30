package questclaim

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/questx-lab/backend/pkg/errorx"
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

func (v *visitLinkValidator) Validate(router.Context, string) (bool, error) {
	return true, nil
}

// Text Validator
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

func (v *textValidator) Validate(ctx router.Context, input string) (bool, error) {
	if !v.AutoValidate {
		return false, errorx.New(errorx.NeedManualReview, "This quest is not auto validated")
	}

	return v.Answer == input, nil
}
