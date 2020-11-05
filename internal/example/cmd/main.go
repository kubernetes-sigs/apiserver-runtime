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

package main

import (
	"sigs.k8s.io/apiserver-runtime/internal/example/v1alpha1"
	"sigs.k8s.io/apiserver-runtime/internal/example/v1beta1"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

func main() {
	var _ resource.Object = &v1alpha1.ExampleResource{}
	var _ resource.Object = &v1beta1.ExampleResource{}

	cmd, err := builder.APIServer.
		// v1alpha1 will be the storage version because it was registered first
		WithResource(&v1alpha1.ExampleResource{}).
		// v1beta1 objects will be converted to v1alpha1 versions before being stored
		WithResource(&v1beta1.ExampleResource{}).
		// OpenAPI definitions are optional for an apiserver, unless you need the openapi
		// functionalities for some cases.
		// WithOpenAPIDefinitions("example", "v0.0.0", openapi.GetOpenAPIDefinitions).
		// Allows you running unsecured apiserver locally.
		WithLocalDebugExtension().
		Build()
	if err != nil {
		panic(err)
	}
	// Call Execute on cmd
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
