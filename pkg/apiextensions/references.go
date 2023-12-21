/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package apiextensions provides helpers structs which can be used anywhere
// +kubebuilder:object:generate:=true
package apiextensions

import "k8s.io/apimachinery/pkg/types"

// ObjectReference is namespaced reference to an object
type ObjectReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (o *ObjectReference) AsNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
}

// ObjectFieldReference is namespaced reference to the object field
type ObjectFieldReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

func (o *ObjectFieldReference) AsNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
}
