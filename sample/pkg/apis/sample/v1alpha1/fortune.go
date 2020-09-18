package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Fortune defines the schema for the "fortunes" resource.
type Fortune struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Value string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
}

type fortuneTableConverter struct {
	metav1.Table
	i int
}

func (c *fortuneTableConverter) convert(obj runtime.Object) error {
	f := obj.(*Fortune)
	value := strings.ReplaceAll(f.Value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.TrimSpace(value)
	if len(value) > 60 {
		value = value[:57] + "..."
	}
	c.Rows = append(c.Rows, metav1.TableRow{
		Cells:  []interface{}{fmt.Sprintf("%d", c.i), value},
		Object: runtime.RawExtension{Object: obj},
	})
	c.i++
	return nil
}

func (c *fortuneTableConverter) setOptions(tableOptions runtime.Object) {
	if opt, ok := tableOptions.(*metav1.TableOptions); !ok || !opt.NoHeaders {
		c.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "Number", Type: "integer", Format: "name", Description: "fortune count"},
			{Name: "Fortune", Type: "string", Format: "name", Description: "random fortune"},
		}
	}
}

// ConvertToTable converts a single resource into a Table
func (f Fortune) convertToTable(_ context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	c := &fortuneTableConverter{}
	if err := c.convert(&f); err != nil {
		return nil, err
	}
	c.ResourceVersion = f.GetResourceVersion()
	c.SelfLink = f.GetSelfLink()
	c.setOptions(tableOptions)
	return &c.Table, nil
}

// ConvertToTable converts a list of resources into a table
func (f FortuneList) convertToTable(_ context.Context, tableOptions runtime.Object) (*metav1.Table, error) {
	c := &fortuneTableConverter{}
	if err := meta.EachListItem(&f, c.convert); err != nil {
		return nil, err
	}
	c.setOptions(tableOptions)
	c.ResourceVersion = f.GetResourceVersion()
	c.SelfLink = f.GetSelfLink()
	c.Continue = f.GetContinue()
	c.RemainingItemCount = f.GetRemainingItemCount()
	return &c.Table, nil
}

// GetGroupVersionResource returns a GroupVersionResource with "fortunes" as the resource.
// GetGroupVersionResource implements resource.Object
func (Fortune) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource("fortunes")
}

// GetObjectMeta implements resource.Object
func (f *Fortune) GetObjectMeta() *metav1.ObjectMeta {
	return &f.ObjectMeta
}

// IsStorageVersion returns true -- v1alpha1.Fortune is used as the internal version.
// IsStorageVersion implements resource.Object.
func (Fortune) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true to indicate Fortune is a namespaced resource.
// NamespaceScoped implements resource.Object.
func (Fortune) NamespaceScoped() bool {
	return true
}

// New implements resource.Object
func (Fortune) New() runtime.Object {
	return &Fortune{}
}

// NewList implements resource.Object
func (Fortune) NewList() runtime.Object {
	return &FortuneList{}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FortuneList is a list of Fortune objects.
type FortuneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Fortune `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// GetListMeta returns the ListMeta
func (f *FortuneList) GetListMeta() *metav1.ListMeta {
	return &f.ListMeta
}
