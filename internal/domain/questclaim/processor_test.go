package questclaim

import (
	"context"
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
			wantErr: errorx.New(errorx.NotFound, "Not found link"),
		},
		{
			name:    "invalid link",
			args:    args{data: map[string]any{"link": "http//example"}},
			want:    nil,
			wantErr: errorx.New(errorx.BadRequest, "Invalid link"),
		},
		{
			name:    "no link field",
			args:    args{data: map[string]any{"link-foo": "http://example.com"}},
			want:    nil,
			wantErr: errorx.New(errorx.NotFound, "Not found link"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newVisitLinkProcessor(testutil.MockContext(), tt.args.data, true)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(got, tt.want), "%v != %v", got, tt.want)
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
			wantErr: errorx.New(errorx.BadRequest, "Invalid twitter handle url"),
		},
		{
			name:    "invalid account url",
			args:    args{data: map[string]any{"twitter_handle": "invalid"}},
			want:    nil,
			wantErr: errorx.New(errorx.BadRequest, "Invalid twitter handle url"),
		},
		{
			name:    "no account url field",
			args:    args{data: map[string]any{"foo": "http://twitter.com/abc"}},
			want:    nil,
			wantErr: errorx.New(errorx.BadRequest, "Invalid twitter handle url"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTwitterFollowProcessor(
				testutil.MockContext(),
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
				require.Equal(t, tt.wantErr, err)
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
		submissionData string
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
			args:   args{submissionData: "any"},
			want:   NeedManualReview,
		},
		{
			name:   "happy case with auto validate",
			fields: fields{AutoValidate: true, Answer: "foo"},
			args:   args{submissionData: "foo"},
			want:   Accepted,
		},
		{
			name:   "wrong answer with auto validate",
			fields: fields{AutoValidate: true, Answer: "foo"},
			args:   args{submissionData: "bar"},
			want:   Rejected.WithMessage("Wrong answer"),
		},
		{
			name:   "wrong answer with no auto validate",
			fields: fields{AutoValidate: false, Answer: "foo"},
			args:   args{submissionData: "bar"},
			want:   NeedManualReview,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &textProcessor{
				AutoValidate: tt.fields.AutoValidate,
				Answer:       tt.fields.Answer,
			}

			got, err := v.GetActionForClaim(testutil.MockContext(), tt.args.submissionData)
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
	type args struct {
		data           map[string]any
		submissionData string
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
					"quizzes": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answers":  []string{"option 1"},
						},
						{
							"question": "question 2",
							"options":  []string{"option A", "option B"},
							"answers":  []string{"option B"},
						},
					},
				},
				submissionData: `{"answers": ["option 1", "option B"]}`,
			},
			want: Accepted,
		},
		{
			name: "happy case with rejected",
			args: args{
				data: map[string]any{
					"quizzes": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answers":  []string{"option 1"},
						},
						{
							"question": "question 2",
							"options":  []string{"option A", "option B"},
							"answers":  []string{"option B"},
						},
					},
				},
				submissionData: `{"answers": ["option 1", "option A"]}`,
			},
			want: Rejected.WithMessage("Wrong answer at quiz 2"),
		},
		{
			name: "invalid answer when new quiz",
			args: args{
				data: map[string]any{
					"quizzes": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answers":  []string{"option B"},
						},
					},
				},
			},
			wantNewErr: errorx.New(errorx.NotFound, "Not found the answer in options"),
		},
		{
			name: "invalid len of answers",
			args: args{
				data: map[string]any{
					"quizzes": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answers":  []string{"option 1"},
						},
					},
				},
				submissionData: `{"answers": ["option 1", "option 2"]}`,
			},
			wantGetActionErr: errorx.New(errorx.BadRequest, "Invalid number of answers"),
		},
		{
			name: "multiple choices of answer",
			args: args{
				data: map[string]any{
					"quizzes": []map[string]any{
						{
							"question": "question 1",
							"options":  []string{"option 1", "option 2"},
							"answers":  []string{"option 1", "option 2"},
						},
					},
				},
				submissionData: `{"answers": ["option 1"]}`,
			},
			want: Accepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := newQuizProcessor(testutil.MockContext(), tt.args.data, true)
			if tt.wantNewErr != nil {
				require.Equal(t, err, tt.wantNewErr)
				return
			} else {
				require.NoError(t, err)
			}

			got, err := v.GetActionForClaim(testutil.MockContext(), tt.args.submissionData)
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
