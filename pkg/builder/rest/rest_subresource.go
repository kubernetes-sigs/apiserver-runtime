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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// NewSubResourceWithStrategy returns a new etcd backed request handler for subresource using the provided Strategy.
func NewSubResourceWithStrategy(parent resource.Object, subResource resource.SubResource, s Strategy) ResourceHandlerProvider {
	return func(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (rest.Storage, error) {
		fullResourceName := parent.GetGroupVersionResource().Resource + "/" + subResource.SubResourceName()
		gvr := parent.GetGroupVersionResource().GroupVersion().WithResource(fullResourceName)
		return newStore(
			scheme,
			parent.New,
			parent.NewList,
			gvr,
			s,
			optsGetter,
			nil)
	}
}
