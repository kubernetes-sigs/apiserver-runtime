// Package mysql provides mysql storage related utilities.
package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/k3s-io/kine/pkg/endpoint"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/util/flowcontrol/request"
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
	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
		username,
		password,
		host,
		port,
		database)

	return func(scheme *runtime.Scheme, s *genericregistry.Store, options *generic.StoreOptions) {
		options.RESTOptions = &kineProxiedRESTOptionsGetter{
			scheme:         scheme,
			dsn:            dsn,
			groupVersioner: s.StorageVersioner,
		}
	}
}

type kineProxiedRESTOptionsGetter struct {
	scheme         *runtime.Scheme
	dsn            string
	groupVersioner runtime.GroupVersioner
}

// GetRESTOptions implements RESTOptionsGetter interface.
func (g *kineProxiedRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	etcdConfig, err := endpoint.Listen(context.TODO(), endpoint.Config{
		Endpoint: g.dsn,
	})
	if err != nil {
		return generic.RESTOptions{}, err
	}
	s := json.NewSerializer(json.DefaultMetaFactory, g.scheme, g.scheme, false)
	codec := serializer.NewCodecFactory(g.scheme).
		CodecForVersions(s, s, g.groupVersioner, g.groupVersioner)
	restOptions := generic.RESTOptions{
		ResourcePrefix:            resource.String(),
		Decorator:                 genericregistry.StorageWithCacher(),
		EnableGarbageCollection:   true,
		DeleteCollectionWorkers:   1,
		CountMetricPollPeriod:     time.Minute,
		StorageObjectCountTracker: request.NewStorageObjectCountTracker(context.Background().Done()),
		StorageConfig: &storagebackend.ConfigForResource{
			GroupResource: resource,
			Config: storagebackend.Config{
				Prefix: "/kine/",
				Codec:  codec,
				Transport: storagebackend.TransportConfig{
					ServerList:    etcdConfig.Endpoints,
					TrustedCAFile: etcdConfig.TLSConfig.CAFile,
					CertFile:      etcdConfig.TLSConfig.CertFile,
					KeyFile:       etcdConfig.TLSConfig.KeyFile,
				},
			},
		},
	}
	return restOptions, nil
}
