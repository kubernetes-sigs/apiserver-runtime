package builder

import "sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/cmd/server"

// SetDelegateAuthOptional makes delegated authentication and authorization optional, otherwise
// the apiserver won't failing upon missing delegated auth configurations.
func (a *Server) SetDelegateAuthOptional() *Server {
	server.ServerOptionsFns = append(server.ServerOptionsFns, func(o *ServerOptions) *ServerOptions {
		o.RecommendedOptions.Authentication.RemoteKubeConfigFileOptional = true
		o.RecommendedOptions.Authorization.RemoteKubeConfigFileOptional = true
		return o
	})
	return a
}

// DisableAuthorization disables delegated authentication and authorization
func (a *Server) DisableAuthorization() *Server {
	server.ServerOptionsFns = append(server.ServerOptionsFns, func(o *ServerOptions) *ServerOptions {
		o.RecommendedOptions.Authorization = nil
		return o
	})
	return a
}
