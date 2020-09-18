package v1alpha1

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcerest"
)

// Rest functions

var _ resourcerest.Getter = &Fortune{}
var _ resourcerest.Lister = &Fortune{}
var _ resourcerest.TableConvertor = &Fortune{}

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
