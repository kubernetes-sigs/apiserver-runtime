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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if getter, isGetter := p.Storage.(rest.Getter); isGetter {
		return parentPlumbedStorageProvider{delegate: getter}, nil
	}
	return p.Storage, nil
}

// ParentStaticHandlerProvider returns itself as the request handler, but with the parent
// storage plumbed in the context.
type ParentStaticHandlerProvider struct {
	rest.Storage
}

// Get returns itself as the handler
func (p ParentStaticHandlerProvider) Get(s *runtime.Scheme, g generic.RESTOptionsGetter) (rest.Storage, error) {
	if getter, isGetter := p.Storage.(rest.Getter); isGetter {
		return parentPlumbedStorageProvider{delegate: getter}, nil
	}
	return p.Storage, nil
}

var _ rest.Getter = &parentPlumbedStorageProvider{}

type parentPlumbedStorageProvider struct {
	delegate rest.Getter
}

func (p parentPlumbedStorageProvider) New() runtime.Object {
	return p.delegate.(rest.Storage).New()
}

func (p parentPlumbedStorageProvider) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	return p.delegate.Get(contextutil.WithParentStorage(ctx, p.delegate.(rest.Storage)), name, options)
}
