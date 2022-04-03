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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// New returns a new etcd backed request handler for the resource.
func New(obj resource.Object) ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (rest.Storage, error) {
		gvr := obj.GetGroupVersionResource()
		s := &DefaultStrategy{
			Object:         obj,
			ObjectTyper:    scheme,
			TableConvertor: rest.NewDefaultTableConvertor(gvr.GroupResource()),
		}
		return newStore(scheme, obj.New, obj.NewList, gvr, s, optsGetter, nil)
	}
}

// NewWithStrategy returns a new etcd backed request handler using the provided Strategy.
func NewWithStrategy(obj resource.Object, s Strategy) ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (rest.Storage, error) {
		gvr := obj.GetGroupVersionResource()
		return newStore(scheme, obj.New, obj.NewList, gvr, s, optsGetter, nil)
	}
}

// StoreFn defines a function which modifies the Store before it is initialized.
type StoreFn func(*runtime.Scheme, *genericregistry.Store, *generic.StoreOptions)

// NewWithFn returns a new etcd backed request handler, applying the StoreFn to the Store.
func NewWithFn(obj resource.Object, fn StoreFn) ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (rest.Storage, error) {
		gvr := obj.GetGroupVersionResource()
		s := &DefaultStrategy{
			Object:         obj,
			ObjectTyper:    scheme,
			TableConvertor: rest.NewDefaultTableConvertor(gvr.GroupResource()),
		}
		return newStore(scheme, obj.New, obj.NewList, gvr, s, optsGetter, fn)
	}
}

// newStore returns a RESTStorage object that will work against API services.
func newStore(
	scheme *runtime.Scheme,
	single, list func() runtime.Object,
	gvr schema.GroupVersionResource,
	s Strategy, optsGetter generic.RESTOptionsGetter, fn StoreFn) (*genericregistry.Store, error) {
	store := &genericregistry.Store{
		NewFunc:                  single,
		NewListFunc:              list,
		PredicateFunc:            s.Match,
		DefaultQualifiedResource: gvr.GroupResource(),
		TableConvertor:           s,
		CreateStrategy:           s,
		UpdateStrategy:           s,
		DeleteStrategy:           s,
		StorageVersioner:         gvr.GroupVersion(),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if fn != nil {
		fn(scheme, store, options)
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}
	return store, nil
}

// GetAttrs returns labels.Set, fields.Set, and error in case the given runtime.Object is not a ObjectMetaProvider
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	provider, ok := obj.(resource.Object)
	if !ok {
		return nil, nil, fmt.Errorf("given object of type %T does not have metadata", obj)
	}
	om := provider.GetObjectMeta()
	return om.GetLabels(), SelectableFields(om), nil
}

// SelectableFields returns a field set that represents the object.
func SelectableFields(obj *metav1.ObjectMeta) fields.Set {
	return generic.ObjectMetaFieldsSet(obj, true)
}

// SubResourceStorageFn is a function that returns objects required to register a subresource into an apiserver
// path is the subresource path from the parent (e.g. "scale"), parent is the resource the subresource
// is under (e.g. &v1.Deployment{}), request is the subresource request (e.g. &Scale{}), storage is
// the storage implementation that handles the requests.
// A SubResourceStorageFn can be used with builder.APIServer.WithSubResourceAndStorageProvider(fn())
type SubResourceStorageFn func() (path string, parent resource.Object, request resource.Object, storage ResourceHandlerProvider)

// ResourceStorageFn is a function that returns the objects required to register a resource into an apiserver.
// request is the resource type (e.g. &v1.Deployment{}), storage is the storage implementation that handles
// the requests.
// A ResourceFn can be used with builder.APIServer.WithResourceAndStorageProvider(fn())
type ResourceStorageFn func() (request resource.Object, storage ResourceHandlerProvider)
