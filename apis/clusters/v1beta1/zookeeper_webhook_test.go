package v1beta1

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/instaclustr/operator/pkg/instaclustr/mock/appversionsmock"
	"github.com/instaclustr/operator/pkg/validation"
)

func Test_zookeeperValidator_ValidateUpdate(t *testing.T) {
	api := appversionsmock.NewInstAPI()
	type fields struct {
		API validation.Validation
	}
	type args struct {
		ctx context.Context
		old runtime.Object
		new runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "updating cluster without settings update",
			fields: fields{
				API: api,
			},
			args: args{
				ctx: nil,
				old: &Zookeeper{
					ObjectMeta: v1.ObjectMeta{
						Generation: 1,
					},
					Spec: ZookeeperSpec{
						GenericClusterSpec: GenericClusterSpec{
							Description: "test",
						},
					},
					Status: ZookeeperStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
				new: &Zookeeper{
					ObjectMeta: v1.ObjectMeta{
						Generation: 2,
					},
					Spec: ZookeeperSpec{
						GenericClusterSpec: GenericClusterSpec{
							Description: "test",
						},
					},
					Status: ZookeeperStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					}},
			},
			wantErr: true,
		},
		{
			name: "valid update operation",
			fields: fields{
				API: api,
			},
			args: args{
				ctx: nil,
				old: &Zookeeper{
					ObjectMeta: v1.ObjectMeta{
						Generation: 1,
					},
					Spec: ZookeeperSpec{
						GenericClusterSpec: GenericClusterSpec{
							Description: "old",
						},
					},
					Status: ZookeeperStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
				new: &Zookeeper{
					ObjectMeta: v1.ObjectMeta{
						Generation: 2,
					},
					Spec: ZookeeperSpec{
						GenericClusterSpec: GenericClusterSpec{
							Description: "new",
						},
					},
					Status: ZookeeperStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zv := &zookeeperValidator{
				API: tt.fields.API,
			}
			if err := zv.ValidateUpdate(tt.args.ctx, tt.args.old, tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
