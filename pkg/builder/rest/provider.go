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

package rest

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apiserver"
	contextutil "sigs.k8s.io/apiserver-runtime/pkg/util/context"
)

// ResourceHandlerProvider provides a request handler for a resource
type ResourceHandlerProvider = apiserver.StorageProvider

// StaticHandlerProvider returns itself as the request handler.
type StaticHandlerProvider struct { // TODO: privatize
	rest.Storage
}

// Get returns itself as the handler
func (p StaticHandlerProvider) Get(s *runtime.Scheme, g generic.RESTOptionsGetter) (rest.Storage, error) {
	return p.Storage, nil
}

// ParentStaticHandlerProvider returns itself as the request handler, but with the parent
// storage plumbed in the context.
type ParentStaticHandlerProvider struct {
	rest.Storage
	ParentProvider ResourceHandlerProvider
}

// Get returns itself as the handler
func (p ParentStaticHandlerProvider) Get(s *runtime.Scheme, g generic.RESTOptionsGetter) (rest.Storage, error) {
	parentStorage, err := p.ParentProvider(s, g)
	if err != nil {
		return nil, err
	}
	getter, isGetter := p.Storage.(rest.Getter)
	updater, isUpdater := p.Storage.(rest.Updater)
	switch {
	case isGetter && isUpdater:
		return parentPlumbedStorageGetterUpdaterProvider{
			getter:        getter,
			updater:       updater,
			parentStorage: parentStorage,
		}, nil
	case isGetter:
		return parentPlumbedStorageGetterProvider{
			delegate:      getter,
			parentStorage: parentStorage,
		}, nil
	}
	return p.Storage, nil
}

var _ rest.Getter = &parentPlumbedStorageGetterProvider{}

type parentPlumbedStorageGetterProvider struct {
	delegate      rest.Getter
	parentStorage rest.Storage
}

func (p parentPlumbedStorageGetterProvider) New() runtime.Object {
	return p.parentStorage.New()
}

func (p parentPlumbedStorageGetterProvider) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return p.delegate.Get(contextutil.WithParentStorage(ctx, p.parentStorage), name, options)
}

var _ rest.Getter = &parentPlumbedStorageGetterUpdaterProvider{}
var _ rest.Updater = &parentPlumbedStorageGetterUpdaterProvider{}

type parentPlumbedStorageGetterUpdaterProvider struct {
	getter        rest.Getter
	updater       rest.Updater
	parentStorage rest.Storage
}

func (p parentPlumbedStorageGetterUpdaterProvider) New() runtime.Object {
	return p.parentStorage.New()
}

func (p parentPlumbedStorageGetterUpdaterProvider) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return p.getter.Get(contextutil.WithParentStorage(ctx, p.parentStorage), name, options)
}

func (p parentPlumbedStorageGetterUpdaterProvider) Update(
	ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	return p.updater.Update(
		contextutil.WithParentStorage(ctx, p.parentStorage),
		name,
		objInfo,
		createValidation,
		updateValidation,
		forceAllowCreate,
		options)
}
