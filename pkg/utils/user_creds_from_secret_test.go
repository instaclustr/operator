package utils

import (
	"testing"

	k8sCore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/instaclustr/operator/pkg/models"
)

func TestGetUserCreds(t *testing.T) {
	type args struct {
		secret *k8sCore.Secret
	}
	tests := []struct {
		name         string
		args         args
		wantUsername string
		wantPassword string
		wantErr      bool
	}{
		{
			name: "success",
			args: args{secret: &k8sCore.Secret{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Immutable:  nil,
				Data:       map[string][]byte{models.Username: []byte(models.Username), models.Password: []byte(models.Password)},
				StringData: nil,
				Type:       "",
			}},
			wantUsername: models.Username,
			wantPassword: models.Password,
			wantErr:      false,
		},
		{
			name: "no username",
			args: args{secret: &k8sCore.Secret{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Immutable:  nil,
				Data:       map[string][]byte{models.Password: []byte(models.Password)},
				StringData: nil,
				Type:       "",
			}},
			wantErr: true,
		},
		{
			name: "no password",
			args: args{secret: &k8sCore.Secret{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Immutable:  nil,
				Data:       map[string][]byte{models.Username: []byte(models.Username)},
				StringData: nil,
				Type:       "",
			}},
			wantErr: true,
		},
		{
			name: "empty map",
			args: args{secret: &k8sCore.Secret{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Immutable:  nil,
				Data:       map[string][]byte{},
				StringData: nil,
				Type:       "",
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsername, gotPassword, err := GetUserCreds(tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUsername != tt.wantUsername {
				t.Errorf("GetUserCreds() gotUsername = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("GetUserCreds() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
		})
	}
}
