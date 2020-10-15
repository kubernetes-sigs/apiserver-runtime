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

package builder_test

import (
	"fmt"

	"sigs.k8s.io/apiserver-runtime/internal/example/handler"
	"sigs.k8s.io/apiserver-runtime/internal/example/strategy"
	"sigs.k8s.io/apiserver-runtime/internal/example/v1alpha1"
	"sigs.k8s.io/apiserver-runtime/internal/example/v1beta1"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/generated/openapi"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
)

// Example registers a resource with the apiserver using etcd for storage.
// If ExampleResource implements resource.Defaulter it will be used for defaulting
func Example() {
	var _ resource.Object = &v1alpha1.ExampleResource{}

	cmd, err := builder.APIServer.
		// Definitions should be generated apiserver-runtime-gen:
		// go get sigs.k8s.io/apiserver-runtime/tools/apiserver-runtime-gen and then add
		// `//go:generate apiserver-runtime-gen` to your main package and run `go generate`
		WithOpenAPIDefinitions("example", "v0.0.0", openapi.GetOpenAPIDefinitions).
		WithResource(&v1alpha1.ExampleResource{}).
		Build()
	if err != nil {
		panic(err)
	}
	// Call Execute on cmd
	fmt.Println(cmd)
}

// Registers multiple versions of the same resource with the apiserver, using etcd for storage.
// The storage version is the first one registered (v1alpha1), and alternate versions (v1beta1) are converted to the
// storage version before being stored.
// Requires that conversion functions be registered with the apiserver.Scheme to convert alternate versions
// to/from the storage version.
func ExampleServer_WithResource() {
	var _ resource.Object = &v1alpha1.ExampleResource{}
	var _ resource.Object = &v1beta1.ExampleResource{}

	cmd, err := builder.APIServer.
		// Definitions should be generated apiserver-runtime-gen:
		// go get sigs.k8s.io/apiserver-runtime/tools/apiserver-runtime-gen and then add
		// `//go:generate apiserver-runtime-gen` to your main package and run `go generate`
		WithOpenAPIDefinitions("example", "v0.0.0", openapi.GetOpenAPIDefinitions).
		// v1alpha1 will be the storage version because it was registered first
		WithResource(&v1alpha1.ExampleResource{}).
		// v1beta1 objects will be converted to v1alpha1 versions before being stored
		WithResource(&v1beta1.ExampleResource{}).
		Build()
	if err != nil {
		panic(err)
	}
	// Call Execute on cmd
	fmt.Println(cmd)
}

// Registers a resource with the apiserver using the ExampleStrategy to configure etcd storage.
// Alternate versions (v1beta1) are converted to v1alpha1 versions before being stored using the ExampleStrategy.
func ExampleServer_WithResourceAndStrategy() {
	var _ resource.Object = &v1alpha1.ExampleResource{}
	var _ resource.Object = &v1beta1.ExampleResource{}
	var _ rest.Strategy = &strategy.ExampleStrategy{}

	cmd, err := builder.APIServer.
		// Definitions should be generated apiserver-runtime-gen:
		// go get sigs.k8s.io/apiserver-runtime/tools/apiserver-runtime-gen and then add
		// `//go:generate apiserver-runtime-gen` to your main package and run `go generate`
		WithOpenAPIDefinitions("example", "v0.0.0", openapi.GetOpenAPIDefinitions).
		// v1alpha1 will be the storage version because it was registered first, and objects will be stored
		// using the provided Strategy
		WithResourceAndStrategy(&v1alpha1.ExampleResource{}, strategy.ExampleStrategy{}).
		// v1beta1 objects will be converted to v1alpha1 versions before being stored using the ExampleStrategy
		WithResource(&v1beta1.ExampleResource{}).
		Build()
	if err != nil {
		panic(err)
	}
	// Call Execute on cmd
	fmt.Println(cmd)
}

// Register a resource with the apiserver using the ExampleHandlerProvider to handle requests rather than
// etcd based storage.
// Request handler handles v1alpha1 versions of the resource, and v1beta1 versions are converted to
// v1alpha1 versions before being handled.
func ExampleServer_WithResourceAndHandler() {
	var _ resource.Object = &v1alpha1.ExampleResource{}
	var _ resource.Object = &v1beta1.ExampleResource{}
	var _ rest.ResourceHandlerProvider = handler.ExampleHandlerProvider

	cmd, err := builder.APIServer.
		// Definitions should be generated apiserver-runtime-gen:
		// go get sigs.k8s.io/apiserver-runtime/tools/apiserver-runtime-gen and then add
		// `//go:generate apiserver-runtime-gen` to your main package and run `go generate`
		WithOpenAPIDefinitions("example", "v0.0.0", openapi.GetOpenAPIDefinitions).
		// v1alpha1 will be the storage version because it was registered first
		WithResourceAndHandler(&v1alpha1.ExampleResource{}, handler.ExampleHandlerProvider).
		// v1beta1 objects will be converted to v1alpha1 versions before the ExampleHandler is invoked
		WithResource(&v1beta1.ExampleResource{}).
		Build()
	if err != nil {
		panic(err)
	}
	// Call Execute on cmd
	fmt.Println(cmd)
}
