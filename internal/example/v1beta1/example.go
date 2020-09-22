/*
Copyright 2020 The Kubernetes Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/internal/example/v1alpha1"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

var _ resource.Object = &ExampleResource{}
var _ resource.ObjectList = &ExampleResourceList{}
var _ resource.MultiVersionObject = &ExampleResource{}

type ExampleResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

type ExampleResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ExampleResource `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (e *ExampleResource) DeepCopyObject() runtime.Object {
	// implemented by code generation
	return e
}

func (e *ExampleResource) GetObjectMeta() *v1.ObjectMeta {
	return &e.ObjectMeta
}

func (e *ExampleResource) NamespaceScoped() bool {
	return true
}

func (e *ExampleResource) New() runtime.Object {
	return &ExampleResource{}
}

func (e *ExampleResource) NewList() runtime.Object {
	return &ExampleResourceList{}
}

func (e *ExampleResource) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "example.com", Version: "v1beta1", Resource: "exampleresources"}
}

func (e *ExampleResource) IsStorageVersion() bool {
	return false
}

func (e *ExampleResourceList) GetListMeta() *metav1.ListMeta {
	// implemented by code generation
	return &e.ListMeta
}

func (e *ExampleResourceList) DeepCopyObject() runtime.Object {
	// implemented by code generation
	return e
}

var _ resource.MultiVersionObject = &ExampleResource{}

func (e *ExampleResource) NewStorageVersionObject() runtime.Object {
	return &v1alpha1.ExampleResource{}
}

func (e *ExampleResource) ConvertToStorageVersion(storageObj runtime.Object) error {
	_ = storageObj.(*v1alpha1.ExampleResource)
	// TODO: do v1beta1 -> v1alpha1 conversion
	return nil
}

func (e *ExampleResource) ConvertFromStorageVersion(storageObj runtime.Object) error {
	_ = storageObj.(*v1alpha1.ExampleResource)
	// TODO: do v1alpha1 -> v1beta1 conversion
	return nil
}
