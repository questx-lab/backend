package token

import (
	"testing"

	"github.com/questx-lab/backend/config"
)

func TestVerify(t *testing.T) {
	type args struct {
		token   string
		configs *config.Configs
	}
	validID := "valid-id"
	validConfigs := &config.Configs{
		JwtSecretKey: "abcxy",
		JwtExpiredAt: 30,
	}
	validToken, _ := Generate(validID, validConfigs)

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				token:   validToken,
				configs: validConfigs,
			},
			wantErr: false,
			want:    validID,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Verify(tt.args.token, tt.args.configs)
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
