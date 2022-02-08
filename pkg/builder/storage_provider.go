package builder

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	registryrest "k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/util"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	contextutil "sigs.k8s.io/apiserver-runtime/pkg/util/context"
)

// singletonProvider ensures different versions of the same resource share storage
type singletonProvider struct {
	sync.Once
	Provider rest.ResourceHandlerProvider
	storage  registryrest.Storage
	err      error
}

func (s *singletonProvider) Get(
	scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (registryrest.Storage, error) {
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

func (s *subResourceStorageProvider) Get(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (registryrest.Storage, error) {
	parentStorage, err := s.parentStorageProvider(scheme, optsGetter)
	if err != nil {
		return nil, err
	}

	var subResourceStorage registryrest.Storage
	if s.subResourceStorageProvider != nil {
		subResourceStorage, err = s.subResourceStorageProvider(scheme, optsGetter)
		if err != nil {
			return nil, err
		}
	}

	// status subresource
	if strings.HasSuffix(s.subResourceGVR.Resource, "/status") {
		stdParentStorage, ok := parentStorage.(registryrest.StandardStorage)
		if !ok {
			return nil, fmt.Errorf("parent storageProvider for %v/%v/%v must implement rest.StandardStorage",
				s.subResourceGVR.Group, s.subResourceGVR.Version, s.subResourceGVR.Resource)
		}
		return createStatusSubResourceStorage(stdParentStorage)
	}
	// scale subresource
	if strings.HasSuffix(s.subResourceGVR.Resource, "/scale") {
		getter, ok := parentStorage.(registryrest.Getter)
		if !ok {
			return nil, fmt.Errorf("parent storageProvider for %v/%v/%v must implement rest.Getter",
				s.subResourceGVR.Group, s.subResourceGVR.Version, s.subResourceGVR.Resource)
		}
		updater, ok := parentStorage.(registryrest.Updater)
		if !ok {
			return nil, fmt.Errorf("parent storageProvider for %v/%v/%v must implement rest.Updater",
				s.subResourceGVR.Group, s.subResourceGVR.Version, s.subResourceGVR.Resource)
		}
		return &scaleSubResourceStorage{
			parentStorage:        parentStorage,
			parentStorageGetter:  getter,
			parentStorageUpdater: updater,
		}, nil
	}
	// connector
	connectorSubResource, isConnector := subResourceStorage.(resource.ConnectorSubResource)
	if isConnector {
		getter, ok := parentStorage.(registryrest.Getter)
		if !ok {
			return nil, fmt.Errorf("parent storageProvider for %v/%v/%v must implement rest.Getter",
				s.subResourceGVR.Group, s.subResourceGVR.Version, s.subResourceGVR.Resource)
		}
		return &connectorSubResourceStorage{
			parentStorage:          parentStorage,
			parentStorageGetter:    getter,
			subResourceConstructor: subResourceStorage,
			subResourceConnector:   connectorSubResource,
		}, nil
	}
	// getter & updater
	getterUpdaterSubResource, isGetterUpdater := subResourceStorage.(resource.GetterUpdaterSubResource)
	if isGetterUpdater {
		stdParentStorage, ok := parentStorage.(registryrest.StandardStorage)
		if ok {
			return &commonSubResourceStorage{
				parentStorage:          stdParentStorage,
				subResourceConstructor: subResourceStorage,
				subResourceGetter:      getterUpdaterSubResource,
				subResourceUpdater:     getterUpdaterSubResource,
			}, nil
		}
		klog.Infof("Parent storageProvider for %v/%v/%v must implement rest.StandardStorage",
			s.subResourceGVR.Group, s.subResourceGVR.Version, s.subResourceGVR.Resource)
	}

	// use the subresource storage directly
	return s.subResourceStorageProvider(scheme, optsGetter)
}

func createStatusSubResourceStorage(parentStorage registryrest.StandardStorage) (registryrest.Storage, error) {
	parentStore, ok := parentStorage.(*registry.Store)
	if !ok {
		return nil, fmt.Errorf("parent type implementing ObjectWithStatusSubResource must be a cananical resource")
	}
	statusStore := *parentStore
	statusStore.UpdateStrategy = &statusSubResourceStrategy{RESTUpdateStrategy: parentStore.UpdateStrategy}
	return &statusSubResourceStorage{
		store: &statusStore,
	}, nil
}

// status subresource storage
type statusSubResourceStorage struct {
	store *registry.Store
}

var _ registryrest.Getter = &statusSubResourceStorage{}
var _ registryrest.Updater = &statusSubResourceStorage{}

func (s *statusSubResourceStorage) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	return s.store.Get(ctx, name, options)
}

func (s *statusSubResourceStorage) New() runtime.Object {
	return s.store.New()
}

func (s *statusSubResourceStorage) Update(ctx context.Context,
	name string,
	objInfo registryrest.UpdatedObjectInfo,
	createValidation registryrest.ValidateObjectFunc,
	updateValidation registryrest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *v1.UpdateOptions) (runtime.Object, bool, error) {
	return s.store.Update(ctx, name, objInfo, createValidation, updateValidation, forceAllowCreate, options)
}

var _ registryrest.RESTUpdateStrategy = &statusSubResourceStrategy{}

// StatusSubResourceStrategy defines a default Strategy for the status subresource.
type statusSubResourceStrategy struct {
	registryrest.RESTUpdateStrategy
}

// PrepareForUpdate calls the PrepareForUpdate function on obj if supported, otherwise does nothing.
func (s *statusSubResourceStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	// should panic/fail-fast upon casting failure
	statusObj := obj.(resource.ObjectWithStatusSubResource)
	statusOld := old.(resource.ObjectWithStatusSubResource)
	// only modifies status
	statusObj.GetStatus().CopyTo(statusOld)
	if err := util.DeepCopy(statusOld, statusObj); err != nil {
		utilruntime.HandleError(err)
	}
}

// common subresource storage
type commonSubResourceStorage struct {
	parentStorage          registryrest.StandardStorage
	subResourceConstructor registryrest.Storage
	subResourceGetter      registryrest.Getter
	subResourceUpdater     registryrest.Updater
}

var _ registryrest.Getter = &commonSubResourceStorage{}
var _ registryrest.Updater = &commonSubResourceStorage{}

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
	objInfo registryrest.UpdatedObjectInfo,
	createValidation registryrest.ValidateObjectFunc,
	updateValidation registryrest.ValidateObjectUpdateFunc,
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

// connector subresource storage
type connectorSubResourceStorage struct {
	parentStorage          registryrest.Storage
	parentStorageGetter    registryrest.Getter
	subResourceConstructor registryrest.Storage
	subResourceConnector   registryrest.Connecter
}

var _ registryrest.Storage = &connectorSubResourceStorage{}
var _ registryrest.Connecter = &connectorSubResourceStorage{}

func (c *connectorSubResourceStorage) New() runtime.Object {
	return c.subResourceConstructor.New()
}

func (c *connectorSubResourceStorage) Connect(ctx context.Context, id string, options runtime.Object, r registryrest.Responder) (http.Handler, error) {
	return c.subResourceConnector.Connect(
		contextutil.WithParentStorage(ctx, c.parentStorage),
		id,
		options,
		r)
}

func (c *connectorSubResourceStorage) NewConnectOptions() (runtime.Object, bool, string) {
	return c.subResourceConnector.NewConnectOptions()
}

func (c *connectorSubResourceStorage) ConnectMethods() []string {
	return c.subResourceConnector.ConnectMethods()
}

// scale subresource storage
type scaleSubResourceStorage struct {
	parentStorage        registryrest.Storage
	parentStorageGetter  registryrest.Getter
	parentStorageUpdater registryrest.Updater
}

func (s *scaleSubResourceStorage) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return autoscalingv1.SchemeGroupVersion.WithKind("Scale")
}

var _ registryrest.GroupVersionKindProvider = &scaleSubResourceStorage{}
var _ registryrest.Getter = &scaleSubResourceStorage{}
var _ registryrest.Updater = &scaleSubResourceStorage{}

func (s *scaleSubResourceStorage) New() runtime.Object {
	return &autoscalingv1.Scale{}
}

func (s *scaleSubResourceStorage) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	parentObj, err := s.parentStorageGetter.Get(
		contextutil.WithParentStorage(ctx, s.parentStorage),
		name,
		options)
	if err != nil {
		return nil, err
	}
	scaleParentObj, ok := parentObj.(resource.ObjectWithScaleSubResource)
	if !ok {
		return nil, fmt.Errorf("not a valid parent object, does it implement resource.ObjectWithScaleSubResource interface?")
	}
	return scaleParentObj.GetScale(), nil
}

func (s *scaleSubResourceStorage) Update(ctx context.Context,
	name string,
	objInfo registryrest.UpdatedObjectInfo,
	createValidation registryrest.ValidateObjectFunc,
	updateValidation registryrest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *v1.UpdateOptions) (runtime.Object, bool, error) {
	updatedObj, updated, err := s.parentStorageUpdater.Update(
		contextutil.WithParentStorage(ctx, s.parentStorage),
		name,
		&scaleUpdatedObjectInfo{reqObjInfo: objInfo},
		toScaleCreateValidation(createValidation),
		toScaleUpdateValidation(updateValidation),
		forceAllowCreate,
		options)
	if err != nil {
		return nil, false, err
	}
	return updatedObj.(resource.ObjectWithScaleSubResource).GetScale(), updated, nil
}

var _ registryrest.UpdatedObjectInfo = &scaleUpdatedObjectInfo{}

type scaleUpdatedObjectInfo struct {
	reqObjInfo registryrest.UpdatedObjectInfo
}

func (s *scaleUpdatedObjectInfo) Preconditions() *v1.Preconditions {
	return s.reqObjInfo.Preconditions()
}

func (s *scaleUpdatedObjectInfo) UpdatedObject(ctx context.Context, oldObj runtime.Object) (newObj runtime.Object, err error) {
	oldObjWithScale := oldObj.(resource.ObjectWithScaleSubResource)
	oldScale := oldObjWithScale.GetScale()
	obj, err := s.reqObjInfo.UpdatedObject(ctx, oldScale)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.NewBadRequest("nil update passed to Scale")
	}
	scale, ok := obj.(*autoscalingv1.Scale)
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("wrong object passed to Scale update: %v", obj))
	}
	oldObjWithScale.SetScale(scale)
	if len(scale.ResourceVersion) != 0 {
		// The client provided a resourceVersion precondition.
		// Set that precondition and return any conflict errors to the client.
		oldObjWithScale.GetObjectMeta().ResourceVersion = scale.ResourceVersion
	}
	return oldObjWithScale, nil
}

func toScaleCreateValidation(f registryrest.ValidateObjectFunc) registryrest.ValidateObjectFunc {
	return func(ctx context.Context, obj runtime.Object) error {
		oldObjWithScale := obj.(resource.ObjectWithScaleSubResource)
		return f(ctx, oldObjWithScale.GetScale())
	}
}

func toScaleUpdateValidation(f registryrest.ValidateObjectUpdateFunc) registryrest.ValidateObjectUpdateFunc {
	return func(ctx context.Context, obj, old runtime.Object) error {
		oldObjWithScale := old.(resource.ObjectWithScaleSubResource)
		objWithScale := obj.(resource.ObjectWithScaleSubResource)
		return f(ctx, objWithScale, oldObjWithScale)
	}
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
