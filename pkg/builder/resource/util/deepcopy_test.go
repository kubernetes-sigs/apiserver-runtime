package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ runtime.Object = &foo{}
var _ runtime.Object = &bar{}
var _ runtime.Object = &baz{}

func TestDeepCopy(t *testing.T) {
	t.Run("deepcopy normal resource should work", func(t *testing.T) {
		src := &foo{}
		dst := &foo{}
		assert.NoError(t, DeepCopy(src, dst))
		assert.True(t, src.deepcopied)
	})
	t.Run("deepcopy resources doesn't implement DeepCopyInto should fail", func(t *testing.T) {
		src := &bar{}
		dst := &bar{}
		assert.Error(t, DeepCopy(src, dst))
	})
	t.Run("deepcopy resource implements wrong DeepCopyInto method should fail", func(t *testing.T) {
		src := &baz{}
		dst := &baz{}
		assert.Error(t, DeepCopy(src, dst))
	})
}

type foo struct {
	deepcopied bool
}

func (f *foo) DeepCopyInto(dst *foo) {
	f.deepcopied = true
}

func (f *foo) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (f *foo) DeepCopyObject() runtime.Object {
	panic("implement me")
}

type bar struct {
}

func (b *bar) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (b *bar) DeepCopyObject() runtime.Object {
	panic("implement me")
}

type baz struct {
}

func (b baz) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (b baz) DeepCopyObject() runtime.Object {
	panic("implement me")
}
