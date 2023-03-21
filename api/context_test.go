package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/questx-lab/backend/mocks"
	"github.com/questx-lab/backend/utils/token"
	"github.com/stretchr/testify/mock"
)

func TestUserIDContext(t *testing.T) {
	type args struct {
		tknGenerator token.Generator
		ctx          *Context
	}
	validUserID := "valid-user-id"
	mockGenerator := mocks.NewGenerator(t)
	tkn := "valid-token"
	mockHttpRequest := httptest.NewRequest("GET", "http://example.com/foo", nil)
	cookie := &http.Cookie{
		Name:  AuthCookie,
		Value: tkn,
	}
	mockHttpRequest.AddCookie(cookie)

	tests := []struct {
		setup func()
		name  string
		args  args
		want  string
	}{
		{
			name: "happy case",
			setup: func() {
				mockGenerator.On("Verify", mock.Anything).Return(validUserID, nil)
			},
			args: args{
				tknGenerator: mockGenerator,
				ctx: &Context{
					Context: context.Background(),
					r:       mockHttpRequest,
					w:       httptest.NewRecorder(),
				},
			},
			want: validUserID,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			ImportUserIDToContext(tt.args.tknGenerator)(tt.args.ctx)
			if got := tt.args.ctx.ExtractUserIDFromContext(); got != tt.want {
				t.Errorf("Context.ExtractUserIDFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
