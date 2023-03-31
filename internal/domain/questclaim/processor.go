package questclaim

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/questx-lab/backend/pkg/xcontext"
)

// VisitLink Processor
type visitLinkProcessor struct {
	Link string `json:"link,omitempty"`
}

func newVisitLinkProcessor(ctx xcontext.Context, data string) (*visitLinkProcessor, error) {
	visitLink := visitLinkProcessor{}
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

func (v *visitLinkProcessor) GetActionForClaim(xcontext.Context, string) (ActionForClaim, error) {
	return Accepted, nil
}

// Text Processor
// TODO: Add retry_after when the claimed quest is rejected by auto validate.
type textProcessor struct {
	AutoValidate bool   `json:"auto_validate"`
	Answer       string `json:"answer"`
}

func newTextProcessor(ctx xcontext.Context, data string) (*textProcessor, error) {
	text := textProcessor{}
	err := json.Unmarshal([]byte(data), &text)
	if err != nil {
		return nil, err
	}

	return &text, nil
}

func (v *textProcessor) GetActionForClaim(ctx xcontext.Context, input string) (ActionForClaim, error) {
	if !v.AutoValidate {
		return NeedManualReview, nil
	}

	if v.Answer != input {
		return Rejected, nil
	}

	return Accepted, nil
}
