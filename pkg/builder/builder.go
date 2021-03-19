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
	"flag"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apiserver"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/cmd/server"
)

// APIServer builds an apiserver to server Kubernetes resources and sub resources.
var APIServer = &Server{
	storageProvider: map[schema.GroupResource]*singletonProvider{},
}

// Server builds a new apiserver for a single API group
type Server struct {
	errs                 []error
	storageProvider      map[schema.GroupResource]*singletonProvider
	groupVersions        map[schema.GroupVersion]bool
	orderedGroupVersions []schema.GroupVersion
	schemes              []*runtime.Scheme
	schemeBuilder        runtime.SchemeBuilder
}

// Build returns a Command used to run the apiserver
func (a *Server) Build() (*Command, error) {
	a.schemes = append(a.schemes, apiserver.Scheme)
	a.schemeBuilder.Register(
		func(scheme *runtime.Scheme) error {
			groupVersions := make(map[string]sets.String)
			for gvr := range apiserver.APIs {
				if groupVersions[gvr.Group] == nil {
					groupVersions[gvr.Group] = sets.NewString()
				}
				groupVersions[gvr.Group].Insert(gvr.Version)
			}
			for g, versions := range groupVersions {
				gvs := []schema.GroupVersion{}
				for _, v := range versions.List() {
					gvs = append(gvs, schema.GroupVersion{
						Group:   g,
						Version: v,
					})
				}
				err := scheme.SetVersionPriority(gvs...)
				if err != nil {
					return err
				}
			}
			for i := range a.orderedGroupVersions {
				metav1.AddToGroupVersion(scheme, a.orderedGroupVersions[i])
			}
			return nil
		},
	)
	for i := range a.schemes {
		if err := a.schemeBuilder.AddToScheme(a.schemes[i]); err != nil {
			panic(err)
		}
	}

	if len(a.errs) != 0 {
		return nil, errs{list: a.errs}
	}
	o := server.NewWardleServerOptions(os.Stdout, os.Stderr, a.orderedGroupVersions...)
	cmd := server.NewCommandStartServer(o, genericapiserver.SetupSignalHandler())
	server.ApplyFlagsFns(cmd.Flags())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	return cmd, nil
}

// Execute builds and executes the apiserver Command.
func (a *Server) Execute() error {
	cmd, err := a.Build()
	if err != nil {
		return err
	}
	return cmd.Execute()
}
