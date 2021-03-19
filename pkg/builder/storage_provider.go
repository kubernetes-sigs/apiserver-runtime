package builder

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	regsitryrest "k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	contextutil "sigs.k8s.io/apiserver-runtime/pkg/util/context"
)

// singletonProvider ensures different versions of the same resource share storage
type singletonProvider struct {
	sync.Once
	Provider rest.ResourceHandlerProvider
	storage  regsitryrest.Storage
	err      error
}

func (s *singletonProvider) Get(
	scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (regsitryrest.Storage, error) {
	s.Once.Do(func() {
		s.storage, s.err = s.Provider(scheme, optsGetter)
	})
	return s.storage, s.err
}

type subResourceStorageProvider struct {
	subResourceGVR             schema.GroupVersionResource
	parentStorageProvider      rest.ResourceHandlerProvider
	subResourceStorageProvider rest.ResourceHandlerProvider
}

func (s *subResourceStorageProvider) Get(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (regsitryrest.Storage, error) {
	parentStorage, err := s.parentStorageProvider(scheme, optsGetter)
	if err != nil {
		return nil, err
	}
	stdParentStorage, ok := parentStorage.(regsitryrest.StandardStorage)
	if !ok {
		return nil, fmt.Errorf("parent storageProvider for %v/%v/%v must implement rest.StandardStorage",
			s.subResourceGVR.Group, s.subResourceGVR.Version, s.subResourceGVR.Resource)
	}

	subResourceStorage, err := s.subResourceStorageProvider(scheme, optsGetter)
	if err != nil {
		return nil, err
	}

	// standard
	if stdSubresourceStorage, isStandardStorage := subResourceStorage.(regsitryrest.StandardStorage); isStandardStorage {
		return &commonSubResourceStorage{
			parentStorage:          stdParentStorage,
			subResourceConstructor: subResourceStorage,
			subResourceGetter:      stdSubresourceStorage,
			subResourceUpdater:     stdSubresourceStorage,
		}, nil
	}
	// getter & updater
	getter, isGetter := subResourceStorage.(regsitryrest.Getter)
	updater, isUpdater := subResourceStorage.(regsitryrest.Updater)
	if isGetter && isUpdater {
		return &commonSubResourceStorage{
			parentStorage:          stdParentStorage,
			subResourceConstructor: subResourceStorage,
			subResourceGetter:      getter,
			subResourceUpdater:     updater,
		}, nil
	}
	// connector
	connector, isConnector := subResourceStorage.(regsitryrest.Connecter)
	if isConnector {
		return &connectorSubResourceStorage{
			parentStorage:          stdParentStorage,
			subResourceConstructor: subResourceStorage,
			subResourceConnector:   connector,
		}, nil
	}

	// use the subresource storage directly
	return s.subResourceStorageProvider(scheme, optsGetter)
}

var _ regsitryrest.Getter = &commonSubResourceStorage{}
var _ regsitryrest.Updater = &commonSubResourceStorage{}

type commonSubResourceStorage struct {
	parentStorage          regsitryrest.StandardStorage
	subResourceConstructor regsitryrest.Storage
	subResourceGetter      regsitryrest.Getter
	subResourceUpdater     regsitryrest.Updater
}

func (c *commonSubResourceStorage) New() runtime.Object {
	return c.subResourceConstructor.New()
}

func (c *commonSubResourceStorage) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	return c.subResourceGetter.Get(
		contextutil.WithParentStorage(ctx, c.parentStorage),
		name,
		options)
}

func (c *commonSubResourceStorage) Update(ctx context.Context,
	name string,
	objInfo regsitryrest.UpdatedObjectInfo,
	createValidation regsitryrest.ValidateObjectFunc,
	updateValidation regsitryrest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *v1.UpdateOptions) (runtime.Object, bool, error) {
	return c.subResourceUpdater.Update(
		contextutil.WithParentStorage(ctx, c.parentStorage),
		name,
		objInfo,
		createValidation,
		updateValidation,
		forceAllowCreate,
		options)
}

var _ regsitryrest.Storage = &connectorSubResourceStorage{}
var _ regsitryrest.Connecter = &connectorSubResourceStorage{}

type connectorSubResourceStorage struct {
	parentStorage          regsitryrest.StandardStorage
	subResourceConstructor regsitryrest.Storage
	subResourceConnector   regsitryrest.Connecter
}

func (c *connectorSubResourceStorage) New() runtime.Object {
	return c.subResourceConstructor.New()
}

func (c *connectorSubResourceStorage) Connect(ctx context.Context, id string, options runtime.Object, r regsitryrest.Responder) (http.Handler, error) {
	return c.subResourceConnector.Connect(ctx, id, options, r)
}

func (c *connectorSubResourceStorage) NewConnectOptions() (runtime.Object, bool, string) {
	return c.subResourceConnector.NewConnectOptions()
}

func (c *connectorSubResourceStorage) ConnectMethods() []string {
	return c.subResourceConnector.ConnectMethods()
}

type errs struct {
	list []error
}

func (e errs) Error() string {
	msgs := []string{fmt.Sprintf("%d errors: ", len(e.list))}
	for i := range e.list {
		msgs = append(msgs, e.list[i].Error())
	}
	return strings.Join(msgs, "\n")
}
