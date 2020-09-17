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

package handler

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic"
	regsitryrest "k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/internal/example/v1alpha1"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
)

var _ rest.ResourceHandlerProvider = ExampleHandlerProvider

func ExampleHandlerProvider(s *runtime.Scheme, _ genericregistry.RESTOptionsGetter) (regsitryrest.Storage, error) {
	return &ExampleHandler{
		DefaultStrategy: builder.DefaultStrategy{
			Object:      &v1alpha1.ExampleResource{},
			ObjectTyper: s,
			TableConvertor: regsitryrest.NewDefaultTableConvertor(
				v1alpha1.ExampleResource{}.GetGroupVersionResource().GroupResource()),
		},
	}, nil
}

var _ regsitryrest.Getter = &ExampleHandler{}
var _ regsitryrest.Lister = &ExampleHandler{}
var _ regsitryrest.CreaterUpdater = &ExampleHandler{}

type ExampleHandler struct {
	builder.DefaultStrategy
}

func (e ExampleHandler) Create(ctx context.Context, obj runtime.Object,
	createValidation regsitryrest.ValidateObjectFunc, options *v1.CreateOptions) (runtime.Object, error) {
	panic("implement me")
}

func (e ExampleHandler) Update(
	ctx context.Context, name string, objInfo regsitryrest.UpdatedObjectInfo,
	createValidation regsitryrest.ValidateObjectFunc, updateValidation regsitryrest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *v1.UpdateOptions) (runtime.Object, bool, error) {
	panic("implement me")
}

func (e ExampleHandler) NewList() runtime.Object {
	panic("implement me")
}

func (e ExampleHandler) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	panic("implement me")
}

func (e ExampleHandler) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	panic("implement me")
}

func (e ExampleHandler) New() runtime.Object {
	return &v1alpha1.ExampleResource{}
}
