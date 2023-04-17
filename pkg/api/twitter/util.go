package twitter

import "github.com/questx-lab/backend/pkg/api"

func IsRateLimit(resp api.JSON) bool {
	errs, err := resp.Get("errors")
	if err != nil {
		return false
	}

	aErrs, ok := errs.([]any)
	if !ok {
		return false
	}

	for i := range aErrs {
		if m, ok := aErrs[i].(map[string]any); ok {
			if code, err := api.JSON(m).GetInt("code"); err != nil && code == 88 {
				return true
			}
		}
	}

	return false
}
