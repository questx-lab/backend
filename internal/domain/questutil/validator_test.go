package questutil

import (
	"reflect"
	"testing"

	"github.com/questx-lab/backend/pkg/testutil"
)

func Test_newVisitLinkValidator(t *testing.T) {
	type args struct {
		data string
	}

	tests := []struct {
		name    string
		args    args
		want    *visitLinkValidator
		wantErr bool
	}{
		{
			name:    "happy case",
			args:    args{data: `{"link": "http://example.com"}`},
			want:    &visitLinkValidator{Link: "http://example.com"},
			wantErr: false,
		},
		{
			name:    "empty link",
			args:    args{data: `{"link": ""}`},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid link",
			args:    args{data: `{"link": "http//example"}`},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "no link field",
			args:    args{data: `{"link-foo": "http://example.com"}`},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newVisitLinkValidator(testutil.NewMockContext(), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("newVisitLinkValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newVisitLinkValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_textValidator_Validate(t *testing.T) {
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
		want    bool
		wantErr bool
	}{
		{
			name:    "happy case with no auto validate",
			fields:  fields{AutoValidate: false},
			args:    args{input: "any"},
			want:    false,
			wantErr: true,
		},
		{
			name:    "happy case with auto validate",
			fields:  fields{AutoValidate: true, Answer: "foo"},
			args:    args{input: "foo"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "wrong answer with auto validate",
			fields:  fields{AutoValidate: true, Answer: "foo"},
			args:    args{input: "bar"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "wrong answer with no auto validate",
			fields:  fields{AutoValidate: false, Answer: "foo"},
			args:    args{input: "bar"},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &textValidator{
				AutoValidate: tt.fields.AutoValidate,
				Answer:       tt.fields.Answer,
			}

			got, err := v.Validate(testutil.NewMockContext(), tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("textValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("textValidator.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
