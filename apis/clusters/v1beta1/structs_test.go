package v1beta1_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
)

func TestUserReferencesDiff(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		old             v1beta1.References
		new             v1beta1.References
		expectedAdded   v1beta1.References
		expectedDeleted v1beta1.References
	}{
		{
			name: "nothing changed",
			old: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
			},
			new: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
			},
		},
		{
			name: "added new reference",
			old: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
			},
			new: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
			},
			expectedAdded: v1beta1.References{
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
			},
		},
		{
			name: "deleted old reference",
			old: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
			},
			new: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
			},
			expectedDeleted: v1beta1.References{
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
			},
		},
		{
			name: "both slices are nil",
			old:  nil,
			new:  nil,
		},
		{
			name: "deleting the first out of 3",
			old: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
				{
					Name:      "name3",
					Namespace: "namespace3",
				},
			},
			new: v1beta1.References{
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
				{
					Name:      "name3",
					Namespace: "namespace3",
				},
			},
			expectedDeleted: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
			},
		},
		{
			name: "deleting the first and adding a new one",
			old: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
				{
					Name:      "name3",
					Namespace: "namespace3",
				},
			},
			new: v1beta1.References{
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
				{
					Name:      "name3",
					Namespace: "namespace3",
				},
				{
					Name:      "name4",
					Namespace: "namespace4",
				},
			},
			expectedDeleted: v1beta1.References{
				{
					Name:      "name1",
					Namespace: "namespace1",
				},
			},
			expectedAdded: v1beta1.References{
				{
					Name:      "name4",
					Namespace: "namespace4",
				},
			},
		},
		{
			name: "deleting the whole references",
			old: v1beta1.References{
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
				{
					Name:      "name3",
					Namespace: "namespace3",
				},
				{
					Name:      "name4",
					Namespace: "namespace4",
				},
			},
			expectedDeleted: v1beta1.References{
				{
					Name:      "name2",
					Namespace: "namespace2",
				},
				{
					Name:      "name3",
					Namespace: "namespace3",
				},
				{
					Name:      "name4",
					Namespace: "namespace4",
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			added, deleted := c.old.Diff(c.new)

			if !reflect.DeepEqual(added, c.expectedAdded) || !reflect.DeepEqual(deleted, c.expectedDeleted) {
				t.Errorf("expected added %s, got %s; expected deleted %s, got %s",
					toJson(c.expectedAdded), toJson(added),
					toJson(c.expectedDeleted), toJson(deleted),
				)
			}
		})
	}
}

func toJson(obj any) string {
	b, _ := json.Marshal(obj)
	return string(b)
}
