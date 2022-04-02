// Package mysql provides mysql storage related utilities.
package mysql

import (
	"context"
	"fmt"

	"github.com/k3s-io/kine/pkg/endpoint"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	builderrest "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
)

// NewMysqlStorageProvider replaces underlying persistent layer (which by default is etcd) w/ MySQL.
// An example of storaing example resource to Mysql will be:
//
//     builder.APIServer.
//       WithResourceAndStorage(&v1alpha1.ExampleResource{}, mysql.NewMysqlStorageProvider(
//             "", // mysql host name		e.g. "127.0.0.1"
//             0,  // mysql password 		e.g. 3306
//             "", // mysql username 		e.g. "mysql"
//             "", // mysql password 		e.g. "password"
//             "", // mysql database name 	e.g. "mydb"
//             )).Build()
//
func NewMysqlStorageProvider(host string, port int32, username, password, database string) builderrest.StoreFn {
	return func(s *genericregistry.Store, options *generic.StoreOptions) {
		options.RESTOptions = &kineProxiedRESTOptionsGetter{
			delegate: options.RESTOptions,
		}
	}
}

type kineProxiedRESTOptionsGetter struct {
	delegate generic.RESTOptionsGetter
}

// GetRESTOptions implements RESTOptionsGetter interface.
func (g *kineProxiedRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	restOptions, err := g.delegate.GetRESTOptions(resource)
	if err != nil {
		return generic.RESTOptions{}, err
	}

	if len(restOptions.StorageConfig.Transport.ServerList) != 1 {
		return generic.RESTOptions{}, fmt.Errorf("no valid mysql dsn found")
	}

	etcdConfig, err := endpoint.Listen(context.TODO(), endpoint.Config{
		Endpoint: restOptions.StorageConfig.Transport.ServerList[0],
	})
	if err != nil {
		return generic.RESTOptions{}, err
	}

	restOptions.StorageConfig.Transport.ServerList = etcdConfig.Endpoints
	restOptions.StorageConfig.Transport.TrustedCAFile = etcdConfig.TLSConfig.CAFile
	restOptions.StorageConfig.Transport.CertFile = etcdConfig.TLSConfig.CertFile
	restOptions.StorageConfig.Transport.KeyFile = etcdConfig.TLSConfig.KeyFile
	return restOptions, nil
}
