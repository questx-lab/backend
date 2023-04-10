package questclaim

import (
	"errors"
	"reflect"
	"testing"

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
			wantErr: errors.New("Not found link in validation data"),
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
			wantErr: errors.New("Not found link in validation data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newVisitLinkProcessor(testutil.NewMockContext(), tt.args.data)
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
			got, err := newTwitterFollowProcessor(testutil.NewMockContext(), tt.args.data)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("newTwitterProcessor() = %v, want %v", got, tt.want)
				}
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

			got, err := v.GetActionForClaim(testutil.NewMockContext(), tt.args.input)
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
