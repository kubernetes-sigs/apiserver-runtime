package context

import (
	"context"

	"k8s.io/apiserver/pkg/registry/rest"
)

type parentStorageContextKeyType string

var parentStorageContextKey parentStorageContextKeyType

// WithParentStorage creates a new child context w/ parent storage plumbed
func WithParentStorage(ctx context.Context, storage rest.Storage) context.Context {
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

// GetParentStorageGetter tries getting the get-only parent storage from
// context.
func GetParentStorageGetter(ctx context.Context) (rest.Getter, bool) {
	parentStorage := ctx.Value(parentStorageContextKey)
	if parentStorage == nil {
		return nil, false
	}
	return parentStorage.(rest.Getter), true
}
