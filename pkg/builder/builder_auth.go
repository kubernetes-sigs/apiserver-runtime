package builder

import (
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/cmd/server"
)

var enableAuthorization bool

// DisableAuthorization disables delegated authentication and authorization
func (a *Server) DisableAuthorization() *Server {
	server.ServerOptionsFns = append(server.ServerOptionsFns, func(o *ServerOptions) *ServerOptions {
		if !enableAuthorization {
			o.RecommendedOptions.Authorization = nil
		}
		return o
	})
	server.FlagsFns = append(server.FlagsFns, func(fs *pflag.FlagSet) *pflag.FlagSet {
		fs.BoolVar(&enableAuthorization, "enable-authorization", false,
			"Enabling authorization will check if the incoming authenticated requests "+
				"have sufficient permission for the requesting target. Deploying the apiserver "+
				"inside a kubernetes cluster will delegate the authorization to the hosting "+
				"kube-apiserver, otherwise specify `--authorization-kubeconfig` to explicitly "+
				"set a kube-apiserver to talk to.")
		return fs
	})
	return a
}

var enablesLocalStandaloneDebugging bool

// WithLocalDebugExtension adds an optional local-debug mode to the apiserver so that it can be tested
// locally without involving a complete kubernetes cluster. A flag named "--standalone-debug-mode" will
// also be added the binary which forcily requires "--bind-address" to be "127.0.0.1" in order to avoid
// security issues.
func (a *Server) WithLocalDebugExtension() *Server {
	server.ServerOptionsFns = append(server.ServerOptionsFns, func(options *ServerOptions) *ServerOptions {
		secureBindingAddr := options.RecommendedOptions.SecureServing.BindAddress.String()
		if enablesLocalStandaloneDebugging {
			if secureBindingAddr != "127.0.0.1" {
				klog.Fatal(`--bind-address must be "127.0.0.1" if --standalone-debug-mode is set`)
			}
			options.RecommendedOptions.Authorization = nil
			options.RecommendedOptions.CoreAPI = nil
			options.RecommendedOptions.Admission = nil
		}
		return options
	})
	server.FlagsFns = append(server.FlagsFns, func(fs *pflag.FlagSet) *pflag.FlagSet {
		fs.BoolVar(&enablesLocalStandaloneDebugging, "standalone-debug-mode", false,
			"Under the local-debug mode the apiserver will allow all access to its resources without "+
				"authorizing the requests, this flag is only intended for debugging in your workstation "+
				"and the apiserver will be crashing if its binding address is not 127.0.0.1.")
		return fs
	})
	server.ServerOptionsFns = append(server.ServerOptionsFns, func(o *ServerOptions) *ServerOptions {
		o.RecommendedOptions.Authentication.RemoteKubeConfigFileOptional = true
		return o
	})
	return a
}
