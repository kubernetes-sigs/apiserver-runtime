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

package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

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

func (e *ExampleResource) GetObjectMeta() *metav1.ObjectMeta {
	return &e.ObjectMeta
}

func (e ExampleResource) NamespaceScoped() bool {
	return true
}

func (e ExampleResource) New() runtime.Object {
	return &ExampleResource{}
}

func (e ExampleResource) NewList() runtime.Object {
	return &ExampleResourceList{}
}

func (e ExampleResource) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "example.com", Version: "v1alpha1", Resource: "exampleresources"}
}

func (e ExampleResource) IsStorageVersion() bool {
	return true
}

func (e *ExampleResourceList) DeepCopyObject() runtime.Object {
	// implemented by code generation
	return e
}

var _ resourcestrategy.TableConverter = &ExampleResource{}
var _ resourcestrategy.TableConverter = &ExampleResourceList{}

var (
	definitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: "the name of the cluster"},
	}
)

func (in *ExampleResource) ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	return &metav1.Table{
		ColumnDefinitions: definitions,
		Rows:              []metav1.TableRow{printResource(in)},
	}, nil
}

func (in *ExampleResourceList) ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	t := &metav1.Table{
		ColumnDefinitions: definitions,
	}
	for _, c := range in.Items {
		t.Rows = append(t.Rows, printResource(&c))
	}
	return t, nil
}

func printResource(c *ExampleResource) metav1.TableRow {
	name := c.Name
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: c},
	}
	row.Cells = append(row.Cells, name)
	return row
}
