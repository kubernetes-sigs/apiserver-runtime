package builder

import (
	"github.com/spf13/pflag"
	pkgserver "k8s.io/apiserver/pkg/server"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apiserver"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/cmd/server"
)

// WithOptionsFns sets functions to customize the ServerOptions used to create the apiserver
func (a *Server) WithOptionsFns(fns ...func(*ServerOptions) *ServerOptions) *Server {
	server.ServerOptionsFns = append(server.ServerOptionsFns, fns...)
	return a
}

// WithServerFns sets functions to customize the GenericAPIServer
func (a *Server) WithServerFns(fns ...func(server *GenericAPIServer) *GenericAPIServer) *Server {
	apiserver.GenericAPIServerFns = append(apiserver.GenericAPIServerFns, fns...)
	return a
}

// WithConfigFns sets functions to customize the RecommendedConfig
func (a *Server) WithConfigFns(fns ...func(config *pkgserver.RecommendedConfig) *pkgserver.RecommendedConfig) *Server {
	server.RecommendedConfigFns = append(server.RecommendedConfigFns, fns...)
	return a
}

// WithFlagFns sets functions to customize the flags for the compiled binary.
func (a *Server) WithFlagFns(fns ...func(set *pflag.FlagSet) *pflag.FlagSet) *Server {
	server.FlagsFns = append(server.FlagsFns, fns...)
	return a
}
