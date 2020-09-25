package builder

import (
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
