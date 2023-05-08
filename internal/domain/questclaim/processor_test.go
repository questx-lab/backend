package questclaim

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func Test_newVisitLinkProcessor(t *testing.T) {
	type args struct {
		data map[string]any
	}

	tests := []struct {
		name    string
		args    args
		want    *visitLinkProcessor
		wantErr error
	}{
		{
			name:    "happy case",
			args:    args{data: map[string]any{"link": "http://example.com"}},
			want:    &visitLinkProcessor{Link: "http://example.com"},
			wantErr: nil,
		},
		{
			name:    "empty link",
			args:    args{data: map[string]any{"link": ""}},
			want:    nil,
			wantErr: errors.New("not found link in validation data"),
		},
		{
			name:    "invalid link",
			args:    args{data: map[string]any{"link": "http//example"}},
			want:    nil,
			wantErr: errors.New("parse \"http//example\": invalid URI for request"),
		},
		{
			name:    "no link field",
			args:    args{data: map[string]any{"link-foo": "http://example.com"}},
			want:    nil,
			wantErr: errors.New("not found link in validation data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newVisitLinkProcessor(testutil.NewMockContext(), tt.args.data, true)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("newVisitLinkProcessor() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_newTwitterFollowProcessor(t *testing.T) {
	type args struct {
		data map[string]any
	}

	tests := []struct {
		name    string
		args    args
		want    *twitterFollowProcessor
		wantErr error
	}{
		{
			name:    "happy case",
			args:    args{data: map[string]any{"twitter_handle": "https://twitter.com/abc"}},
			want:    &twitterFollowProcessor{TwitterHandle: "https://twitter.com/abc"},
			wantErr: nil,
		},
		{
			name:    "empty account url",
			args:    args{data: map[string]any{"twitter_handle": ""}},
			want:    nil,
			wantErr: errors.New("parse \"\": empty url"),
		},
		{
			name:    "invalid account url",
			args:    args{data: map[string]any{"twitter_handle": "invalid"}},
			want:    nil,
			wantErr: errors.New("parse \"invalid\": invalid URI for request"),
		},
		{
			name:    "no account url field",
			args:    args{data: map[string]any{"foo": "http://twitter.com/abc"}},
			want:    nil,
			wantErr: errors.New("parse \"\": empty url"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTwitterFollowProcessor(
				testutil.NewMockContext(),
				Factory{
					twitterEndpoint: &testutil.MockTwitterEndpoint{
						GetUserFunc: func(ctx context.Context, s string) (twitter.User, error) {
							return twitter.User{}, nil
						},
					},
				},
				tt.args.data, true,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, tt.want.TwitterHandle, got.TwitterHandle)
			}
		})
	}
}

func Test_textProcessor_GetActionForClaim(t *testing.T) {
	type fields struct {
		AutoValidate bool
		Answer       string
	}
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ActionForClaim
		wantErr error
	}{
		{
			name:   "happy case with no auto validate",
			fields: fields{AutoValidate: false},
			args:   args{input: "any"},
			want:   NeedManualReview,
		},
		{
			name:   "happy case with auto validate",
			fields: fields{AutoValidate: true, Answer: "foo"},
			args:   args{input: "foo"},
			want:   Accepted,
		},
		{
			name:   "wrong answer with auto validate",
			fields: fields{AutoValidate: true, Answer: "foo"},
			args:   args{input: "bar"},
			want:   Rejected,
		},
		{
			name:   "wrong answer with no auto validate",
			fields: fields{AutoValidate: false, Answer: "foo"},
			args:   args{input: "bar"},
			want:   NeedManualReview,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &textProcessor{
				AutoValidate: tt.fields.AutoValidate,
				Answer:       tt.fields.Answer,
			}

			got, err := v.GetActionForClaim(testutil.NewMockContext(), nil, tt.args.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}

func Test_quizProcessor(t *testing.T) {
	type fields struct {
		AutoValidate bool
		Answer       string
	}
	type args struct {
		data  map[string]any
		input string
	}
	tests := []struct {
		name             string
		args             args
		want             ActionForClaim
		wantNewErr       error
		wantGetActionErr error
	}{
		{
			name: "happy case with accepted",
			args: args{
				data: map[string]any{
					"quizs": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answer":   "option 1",
						},
						{
							"question": "question 2",
							"options":  []string{"option A", "option B"},
							"answer":   "option B",
						},
					},
				},
				input: `{"answers": ["option 1", "option B"]}`,
			},
			want: Accepted,
		},
		{
			name: "happy case with rejected",
			args: args{
				data: map[string]any{
					"quizs": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answer":   "option 1",
						},
						{
							"question": "question 2",
							"options":  []string{"option A", "option B"},
							"answer":   "option B",
						},
					},
				},
				input: `{"answers": ["option 1", "option A"]}`,
			},
			want: Rejected,
		},
		{
			name: "invalid answer when new quiz",
			args: args{
				data: map[string]any{
					"quizs": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answer":   "option B",
						},
					},
				},
			},
			wantNewErr: errors.New("not found the answer in options"),
		},
		{
			name: "invalid len of answers",
			args: args{
				data: map[string]any{
					"quizs": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answer":   "option 1",
						},
					},
				},
				input: `{"answers": ["option 1", "option 2"]}`,
			},
			wantGetActionErr: errorx.New(errorx.BadRequest, "Invalid number of answers"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := newQuizProcessor(testutil.NewMockContext(), tt.args.data, true)
			if tt.wantNewErr != nil {
				require.Equal(t, err.Error(), tt.wantNewErr.Error())
				return
			} else {
				require.NoError(t, err)
			}

			got, err := v.GetActionForClaim(testutil.NewMockContext(), nil, tt.args.input)
			if tt.wantGetActionErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, tt.wantGetActionErr, err)
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tt.want, got), "%v != %v", tt.want, got)
			}
		})
	}
}
