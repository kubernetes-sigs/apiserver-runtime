package context

import (
	"context"

	"k8s.io/apiserver/pkg/registry/rest"
)

type parentStorageContextKeyType string

var parentStorageContextKey parentStorageContextKeyType = ""

// WithParentStorage creates a new child context w/ parent storage plumbed
func WithParentStorage(ctx context.Context, storage rest.StandardStorage) context.Context {
	return context.WithValue(ctx, parentStorageContextKey, storage)
}

// GetParentStorage tries getting the parent storage from context
func GetParentStorage(ctx context.Context) (rest.StandardStorage, bool) {
	parentStorage := ctx.Value(parentStorageContextKey)
	if parentStorage == nil {
		return nil, false
	}
	return parentStorage.(rest.StandardStorage), true
}
