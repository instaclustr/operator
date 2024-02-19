package v1beta1

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKafkaACL_validateCreate(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       KafkaACLSpec
		Status     KafkaACLStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty acls field",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid principal",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal: "invalid",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid prefix in user query",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal:      "User:test",
						PermissionType: "DENY",
						PatternType:    "LITERAL",
						Operation:      "ALL",
						ResourceType:   "CLUSTER",
					}},
					UserQuery: "User:test",
				},
			},
			wantErr: true,
		},
		{
			name: "valid spec",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal:      "User:test",
						PermissionType: "DENY",
						PatternType:    "LITERAL",
						Operation:      "ALL",
						ResourceType:   "CLUSTER",
					}},
					UserQuery: "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kacl := &KafkaACL{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if err := kacl.validateCreate(); (err != nil) != tt.wantErr {
				t.Errorf("validateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	type args struct {
		acl ACL
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid PermissionType",
			args: args{
				acl: ACL{
					PermissionType: "test",
					PatternType:    "LITERAL",
					Operation:      "ALL",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: true,
		},
		{
			name: "valid PermissionType",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "LITERAL",
					Operation:      "ALL",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid PatternType",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "test",
					Operation:      "ALL",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: true,
		},
		{
			name: "valid PatternType",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "LITERAL",
					Operation:      "ALL",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid Operation",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "LITERAL",
					Operation:      "test",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: true,
		},
		{
			name: "valid Operation",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "LITERAL",
					Operation:      "ALL",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid ResourceType",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "LITERAL",
					Operation:      "ALL",
					ResourceType:   "test",
				},
			},
			wantErr: true,
		},
		{
			name: "valid ResourceType",
			args: args{
				acl: ACL{
					PermissionType: "DENY",
					PatternType:    "LITERAL",
					Operation:      "ALL",
					ResourceType:   "CLUSTER",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validate(tt.args.acl); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKafkaACL_validateUpdate(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       KafkaACLSpec
		Status     KafkaACLStatus
	}
	type args struct {
		oldKafkaACL *KafkaACL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "empty acls",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "change clusterID",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal:      "User:test",
						PermissionType: "DENY",
						PatternType:    "LITERAL",
						Operation:      "ALL",
						ResourceType:   "CLUSTER",
					}},
					ClusterID: "new",
				},
			},
			args: args{
				oldKafkaACL: &KafkaACL{
					Spec: KafkaACLSpec{
						ACLs: []ACL{{
							Principal:      "User:test",
							PermissionType: "DENY",
							PatternType:    "LITERAL",
							Operation:      "ALL",
							ResourceType:   "CLUSTER",
						}},
						ClusterID: "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "change userQuery",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal:      "User:test",
						PermissionType: "DENY",
						PatternType:    "LITERAL",
						Operation:      "ALL",
						ResourceType:   "CLUSTER",
					}},
					ClusterID: "test",
					UserQuery: "new",
				},
			},
			args: args{
				oldKafkaACL: &KafkaACL{
					Spec: KafkaACLSpec{
						ACLs: []ACL{{
							Principal:      "User:test",
							PermissionType: "DENY",
							PatternType:    "LITERAL",
							Operation:      "ALL",
							ResourceType:   "CLUSTER",
						}},
						ClusterID: "test",
						UserQuery: "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "different principal and userQuery",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal:      "User:test",
						PermissionType: "DENY",
						PatternType:    "LITERAL",
						Operation:      "ALL",
						ResourceType:   "CLUSTER",
					}},
					ClusterID: "test",
					UserQuery: "diff",
				},
			},
			args: args{
				oldKafkaACL: &KafkaACL{
					Spec: KafkaACLSpec{
						ACLs: []ACL{{
							Principal:      "User:test",
							PermissionType: "DENY",
							PatternType:    "LITERAL",
							Operation:      "ALL",
							ResourceType:   "CLUSTER",
						}},
						ClusterID: "test",
						UserQuery: "diff",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid update",
			fields: fields{
				Spec: KafkaACLSpec{
					ACLs: []ACL{{
						Principal:      "User:test",
						PermissionType: "DENY",
						PatternType:    "LITERAL",
						Operation:      "ALL",
						ResourceType:   "CLUSTER",
					},
						{
							Principal:      "User:test",
							PermissionType: "DENY",
							PatternType:    "LITERAL",
							Operation:      "ALL",
							ResourceType:   "CLUSTER",
						},
					},
					ClusterID: "test",
					UserQuery: "test",
				},
			},
			args: args{
				oldKafkaACL: &KafkaACL{
					Spec: KafkaACLSpec{
						ACLs: []ACL{{
							Principal:      "User:test",
							PermissionType: "DENY",
							PatternType:    "LITERAL",
							Operation:      "ALL",
							ResourceType:   "CLUSTER",
						}},
						ClusterID: "test",
						UserQuery: "test",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kacl := &KafkaACL{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if err := kacl.validateUpdate(tt.args.oldKafkaACL); (err != nil) != tt.wantErr {
				t.Errorf("validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
