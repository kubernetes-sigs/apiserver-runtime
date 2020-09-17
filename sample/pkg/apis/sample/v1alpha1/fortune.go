package v1alpha1

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcerest"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
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

// Rest functions

var _ resourcerest.Getter = &Fortune{}
var _ resourcerest.Lister = &Fortune{}

// ConvertToTable handles table printing from kubectl get
func (f *Fortune) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	switch o := object.(type) {
	case *Fortune:
		return o.convertToTable(ctx, tableOptions)
	case *FortuneList:
		return o.convertToTable(ctx, tableOptions)
	}
	return nil, fmt.Errorf("unknown type Fortune %T", f)
}

// List implements rest.Lister
func (f *Fortune) List(ctx context.Context, o *internalversion.ListOptions) (runtime.Object, error) {
	parts := strings.SplitN(o.LabelSelector.String(), "=", 2)

	if len(parts) == 1 {
		fl := &FortuneList{}
		// return 5 random fortunes
		for i := 0; i < 5; i++ {
			obj, err := f.Get(ctx, "", &metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			fl.Items = append(fl.Items, *obj.(*Fortune))
		}
		return fl, nil
	}

	fl := &FortuneList{}
	var out []byte
	/* #nosec */
	out, _ = exec.Command("/usr/games/fortune", "-s", "-m", parts[1]).Output()
	values := strings.Split(string(out), "\n%\n")
	for i, fo := range values {
		if i > 5 {
			break
		}
		if strings.TrimSpace(fo) == "" {
			continue
		}
		fl.Items = append(fl.Items, Fortune{Value: strings.TrimSpace(fo)})
	}
	return fl, nil
}

// Get implements rest.Getter
func (f *Fortune) Get(_ context.Context, name string, _ *metav1.GetOptions) (runtime.Object, error) {
	obj := &Fortune{}
	var out []byte
	// fortune exits non-zero on success
	if name == "" {
		out, _ = exec.Command("/usr/games/fortune", "-s").Output()
	} else {
		/* #nosec */
		out, _ = exec.Command("/usr/games/fortune", "-s", "-m", name).Output()
		fortunes := strings.Split(string(out), "\n%\n")
		if len(fortunes) > 0 {
			out = []byte(fortunes[0])
		}
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return nil, errors.NewNotFound(Fortune{}.GetGroupVersionResource().GroupResource(), name)
	}
	obj.Value = strings.TrimSpace(string(out))
	return obj, nil
}
