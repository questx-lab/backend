package token

import (
	"testing"
)

func TestJwtGenerator(t *testing.T) {
	type args struct {
		token        string
		jwtGenerator Generator
	}
	validID := "valid-id"

	validGenerator := NewJWTGenerator("test-token", &Configs{
		JwtSecretKey: "abcxy",
		JwtExpiredAt: 30,
	})

	validToken, _ := validGenerator.Generate(validID)

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				token:        validToken,
				jwtGenerator: validGenerator,
			},
			wantErr: false,
			want:    validID,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.jwtGenerator.Verify(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
