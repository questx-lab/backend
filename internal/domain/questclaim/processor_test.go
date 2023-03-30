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
		data string
	}

	tests := []struct {
		name    string
		args    args
		want    *visitLinkProcessor
		wantErr error
	}{
		{
			name:    "happy case",
			args:    args{data: `{"link": "http://example.com"}`},
			want:    &visitLinkProcessor{Link: "http://example.com"},
			wantErr: nil,
		},
		{
			name:    "empty link",
			args:    args{data: `{"link": ""}`},
			want:    nil,
			wantErr: errors.New("Not found link in validation data"),
		},
		{
			name:    "invalid link",
			args:    args{data: `{"link": "http//example"}`},
			want:    nil,
			wantErr: errors.New("parse \"http//example\": invalid URI for request"),
		},
		{
			name:    "no link field",
			args:    args{data: `{"link-foo": "http://example.com"}`},
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
