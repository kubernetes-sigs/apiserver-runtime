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

package builder

import (
	"strings"

	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/runtime/schema"
	regsitryrest "k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apiserver"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcerest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
)

// WithResource registers the resource with the apiserver.
//
// If no versions of this GroupResource have already been registered, a new default handler will be registered.
// If the object implements rest.Getter, rest.Updater or rest.Creator then the provided object itself will be
// used as the rest handler for the resource type.
//
// If no versions of this GroupResource have already been registered and the object does NOT implement the rest
// interfaces, then a new etcd backed storage will be created for the object and used as the handler.
// The storage will use a DefaultStrategy, which delegates functions to the object if the object implements
// interfaces defined in the "apiserver-runtime/pkg/builder/rest" package.  Otherwise it will provide a default
// behavior.
//
// WithResource will automatically register the "status" subresource if the object implements the
// resource.StatusGetSetter interface.
//
// WithResource will automatically register version-specific defaulting for this GroupVersionResource
// if the object implements the resource.Defaulter interface.
//
// WithResource automatically adds the object and its list type to the known types.  If the object also declares itself
// as the storage version, the object and its list type will be added as storage versions to the SchemeBuilder as well.
// The storage version is the version accepted by the handler.
//
// If another version of the object's GroupResource has already been registered, then the resource will use the
// handler already registered for that version of the GroupResource.  Objects for this version will be converted
// to the object version which the handler accepts before the handler is invoked.
func (a *Server) WithResource(obj resource.Object) *Server {
	gvr := obj.GetGroupVersionResource()
	a.schemeBuilder.Register(resource.AddToScheme(obj))

	// reuse the storage if this resource has already been registered
	if s, found := a.storageProvider[gvr.GroupResource()]; found {
		_ = a.forGroupVersionResource(gvr, s.Get)
		return a
	}

	var parentStorageProvider rest.ResourceHandlerProvider

	defer func() {
		// automatically create status subresource if the object implements the status interface
		a.withSubResourceIfExists(obj, parentStorageProvider)
	}()

	// If the type implements it's own storage, then use that
	switch s := obj.(type) {
	case resourcerest.Creator, resourcerest.Updater, resourcerest.Getter, resourcerest.Lister:
		parentStorageProvider = rest.StaticHandlerProvider{Storage: s.(regsitryrest.Storage)}.Get
	default:
		parentStorageProvider = rest.New(obj)
	}

	_ = a.forGroupVersionResource(gvr, parentStorageProvider)

	return a
}

// WithResourceAndStrategy registers the resource with the apiserver creating a new etcd backed storage
// for the GroupResource using the provided strategy.  In most cases callers should instead use WithResource
// and implement the interfaces defined in "apiserver-runtime/pkg/builder/rest" to control the Strategy.
//
// Note: WithResourceAndHandler should never be called after the GroupResource has already been registered with
// another version.
func (a *Server) WithResourceAndStrategy(obj resource.Object, strategy rest.Strategy) *Server {
	gvr := obj.GetGroupVersionResource()
	a.schemeBuilder.Register(resource.AddToScheme(obj))

	parentStorageProvider := rest.NewWithStrategy(obj, strategy)
	_ = a.forGroupVersionResource(gvr, parentStorageProvider)

	// automatically create status subresource if the object implements the status interface
	a.withSubResourceIfExists(obj, parentStorageProvider)
	return a
}

// WithResourceAndHandler registers a request handler for the resource rather than the default
// etcd backed storage.
//
// Note: WithResourceAndHandler should never be called after the GroupResource has already been registered with
// another version.
//
// Note: WithResourceAndHandler will NOT register the "status" subresource for the resource object.
func (a *Server) WithResourceAndHandler(obj resource.Object, sp rest.ResourceHandlerProvider) *Server {
	gvr := obj.GetGroupVersionResource()
	a.schemeBuilder.Register(resource.AddToScheme(obj))
	return a.forGroupVersionResource(gvr, sp)
}

// WithResourceAndStorage registers the resource with the apiserver, applying fn to the storage for the resource
// before completing it.
//
// May be used to change low-level storage configuration or swap out the storage backend to something other than
// etcd.
//
// Note: WithResourceAndHandler should never be called after the GroupResource has already been registered with
// another version.
func (a *Server) WithResourceAndStorage(obj resource.Object, fn rest.StoreFn) *Server {
	gvr := obj.GetGroupVersionResource()
	a.schemeBuilder.Register(resource.AddToScheme(obj))
	return a.forGroupVersionResource(gvr, rest.NewWithFn(obj, fn))
}

// forGroupVersionResource manually registers storage for a specific resource.
func (a *Server) forGroupVersionResource(
	gvr schema.GroupVersionResource, sp rest.ResourceHandlerProvider) *Server {
	// register the group version
	a.withGroupVersions(gvr.GroupVersion())

	// TODO: make sure folks don't register multiple storageProvider instance for the same group-resource
	// don't replace the existing instance otherwise it will chain wrapped singletonProviders when
	// fetching from the map before calling this function
	if _, found := a.storageProvider[gvr.GroupResource()]; !found {
		a.storageProvider[gvr.GroupResource()] = &singletonProvider{Provider: sp}
	}
	// add the API with its storageProvider
	apiserver.APIs[gvr] = sp
	return a
}

// forGroupVersionSubResource manually registers storageProvider for a specific subresource.
func (a *Server) forGroupVersionSubResource(
	gvr schema.GroupVersionResource, parentProvider rest.ResourceHandlerProvider, subResourceProvider rest.ResourceHandlerProvider) {
	isSubResource := strings.Contains(gvr.Resource, "/")
	if !isSubResource {
		klog.Fatalf("Expected status subresource but received %v/%v/%v", gvr.Group, gvr.Version, gvr.Resource)
	}

	// add the API with its storageProvider for subresource
	apiserver.APIs[gvr] = (&subResourceStorageProvider{
		subResourceGVR:             gvr,
		parentStorageProvider:      parentProvider,
		subResourceStorageProvider: subResourceProvider,
	}).Get
}

// WithSchemeInstallers registers functions to install resource types into the Scheme.
func (a *Server) withGroupVersions(versions ...schema.GroupVersion) *Server {
	if a.groupVersions == nil {
		a.groupVersions = map[schema.GroupVersion]bool{}
	}
	for _, gv := range versions {
		if _, found := a.groupVersions[gv]; found {
			continue
		}
		a.groupVersions[gv] = true
		a.orderedGroupVersions = append(a.orderedGroupVersions, gv)
	}
	return a
}

func (a *Server) withSubResourceIfExists(obj resource.Object, parentStorageProvider rest.ResourceHandlerProvider) {
	parentGVR := obj.GetGroupVersionResource()
	// automatically create status subresource if the object implements the status interface
	if _, ok := obj.(resource.ObjectWithStatusSubResource); ok {
		statusGVR := parentGVR.GroupVersion().WithResource(parentGVR.Resource + "/status")
		a.forGroupVersionSubResource(statusGVR, parentStorageProvider, nil)
	}
	if _, ok := obj.(resource.ObjectWithScaleSubResource); ok {
		subResourceGVR := parentGVR.GroupVersion().WithResource(parentGVR.Resource + "/scale")
		a.forGroupVersionSubResource(subResourceGVR, parentStorageProvider, nil)
	}
	if sgs, ok := obj.(resource.ObjectWithArbitrarySubResource); ok {
		for _, sub := range sgs.GetArbitrarySubResources() {
			sub := sub
			subResourceGVR := parentGVR.GroupVersion().WithResource(parentGVR.Resource + "/" + sub.SubResourceName())
			a.forGroupVersionSubResource(subResourceGVR, parentStorageProvider, rest.ParentStaticHandlerProvider{
				Storage:        sub,
				ParentProvider: parentStorageProvider,
			}.Get)
		}
	}
}
