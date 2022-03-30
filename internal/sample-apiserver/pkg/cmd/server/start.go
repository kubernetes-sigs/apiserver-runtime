/*
Copyright 2016 The Kubernetes Authors.

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

package server

import (
	"fmt"
	"io"
	"net"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/features"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apiserver"
)

// change: apiserver-runtime
//const defaultEtcdPathPrefix = "/registry/wardle.example.com"

// WardleServerOptions contains state for master/api server
type WardleServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions

	StdOut io.Writer
	StdErr io.Writer
}

// NewWardleServerOptions returns a new WardleServerOptions
func NewWardleServerOptions(out, errOut io.Writer, versions ...schema.GroupVersion) *WardleServerOptions {
	// change: apiserver-runtime
	o := &WardleServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			getEctdPath(),
			apiserver.Codecs.LegacyCodec(versions...),
		),

		StdOut: out,
		StdErr: errOut,
	}
	o.RecommendedOptions.Etcd.StorageConfig.EncodeVersioner = schema.GroupVersions(versions)
	return o
}

// NewCommandStartWardleServer provides a CLI handler for 'start master' command
// with a default WardleServerOptions.
func NewCommandStartWardleServer(defaults *WardleServerOptions, stopCh <-chan struct{}) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "Launch a wardle API server",
		Long:  "Launch a wardle API server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunWardleServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)
	utilfeature.DefaultMutableFeatureGate.AddFlag(flags)

	return cmd
}

// Validate validates WardleServerOptions
func (o WardleServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

// Complete fills in fields required to have valid data
func (o *WardleServerOptions) Complete() error {
	// change: apiserver-runtime
	//// register admission plugins
	//banflunder.Register(o.RecommendedOptions.Admission.Plugins)
	//
	//// add admisison plugins to the RecommendedPluginOrder
	//o.RecommendedOptions.Admission.RecommendedPluginOrder = append(o.RecommendedOptions.Admission.RecommendedPluginOrder, "BanFlunder")

	ApplyServerOptionsFns(o)

	return nil
}

// Config returns config for the api server given WardleServerOptions
func (o *WardleServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	// change: allow etcd options to be nil
	// TODO: this should be reverted after rebasing sample-apiserver onto https://github.com/kubernetes/kubernetes/pull/101106
	if o.RecommendedOptions.Etcd != nil {
		o.RecommendedOptions.Etcd.StorageConfig.Paging = utilfeature.DefaultFeatureGate.Enabled(features.APIListChunking)
	}

	// change: apiserver-runtime
	// ExtraAdmissionInitializers set through ApplyServerOptionsFns by appending to ServerOptionsFns
	//
	// o.RecommendedOptions.ExtraAdmissionInitializers = func(c *genericapiserver.RecommendedConfig) ([]admission.PluginInitializer, error) {
	//	 client, err := clientset.NewForConfig(c.LoopbackClientConfig)
	//	 if err != nil {
	//		 return nil, err
	//	 }
	//	 informerFactory := informers.NewSharedInformerFactory(client, c.LoopbackClientConfig.Timeout)
	//	 o.SharedInformerFactory = informerFactory
	//	 return []admission.PluginInitializer{wardleinitializer.New(informerFactory)}, nil
	// }

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)

	// change: apiserver-runtime
	// OpenAPIConfig set through ApplyRecommendedConfigFns by calling SetOpenAPIDefinitions
	//
	// serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(sampleopenapi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(apiserver.Scheme))
	// serverConfig.OpenAPIConfig.Info.Title = "Wardle"
	// serverConfig.OpenAPIConfig.Info.Version = "0.1"

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	serverConfig = ApplyRecommendedConfigFns(serverConfig)

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}
	return config, nil
}

// RunWardleServer starts a new WardleServer given WardleServerOptions
func (o WardleServerOptions) RunWardleServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHookOrDie("start-sample-server-informers", func(context genericapiserver.PostStartHookContext) error {
		if config.GenericConfig.SharedInformerFactory != nil {
			config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
		}
		return nil
	})

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
