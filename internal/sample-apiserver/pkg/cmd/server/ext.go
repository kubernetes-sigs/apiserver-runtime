/*
Copyright 2017 The Kubernetes Authors.

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
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	pkgserver "k8s.io/apiserver/pkg/server"
	openapicommon "k8s.io/kube-openapi/pkg/common"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apiserver"
)

var (
	EtcdPath              string
	RecommendedConfigFns  []func(*pkgserver.RecommendedConfig) *pkgserver.RecommendedConfig
	ServerOptionsFns      []func(server *ServerOptions) *ServerOptions
	FlagsFns              []func(fs *pflag.FlagSet) *pflag.FlagSet
	NewCommandStartServer = NewCommandStartWardleServer
)

type ServerOptions = WardleServerOptions

func ApplyServerOptionsFns(in *ServerOptions) *ServerOptions {
	for i := range ServerOptionsFns {
		in = ServerOptionsFns[i](in)
	}
	return in
}

func ApplyRecommendedConfigFns(in *pkgserver.RecommendedConfig) *pkgserver.RecommendedConfig {
	for i := range RecommendedConfigFns {
		in = RecommendedConfigFns[i](in)
	}
	return in
}

func ApplyFlagsFns(fs *pflag.FlagSet) *pflag.FlagSet {
	for i := range FlagsFns {
		fs = FlagsFns[i](fs)
	}
	return fs
}

func SetOpenAPIDefinitions(name, version string, defs openapicommon.GetOpenAPIDefinitions) {
	RecommendedConfigFns = append(RecommendedConfigFns, func(config *pkgserver.RecommendedConfig) *pkgserver.RecommendedConfig {
		config.OpenAPIConfig = pkgserver.DefaultOpenAPIConfig(defs, openapi.NewDefinitionNamer(apiserver.Scheme))
		config.OpenAPIConfig.Info.Title = name
		config.OpenAPIConfig.Info.Version = version
		return config
	})
}

func getEctdPath() string {
	// TODO: make this configurable
	return "/registry/sample-apiserver"
}
