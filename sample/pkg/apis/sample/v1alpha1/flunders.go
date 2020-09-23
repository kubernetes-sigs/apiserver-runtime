/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"

	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

var _ resource.Object = &Flunder{}
var _ resource.ObjectList = &FlunderList{}
var _ resourcestrategy.Validater = &Flunder{}
var _ resourcestrategy.ValidateUpdater = &Flunder{}

// ReferenceType defines the type of an object reference.
type ReferenceType string

const (
	// FlunderReferenceType is used for Flunder references.
	FlunderReferenceType = ReferenceType("Flunder")
	// FischerReferenceType is used for Fischer references.
	FischerReferenceType = ReferenceType("Fischer")
)

// FlunderSpec is the specification of a Flunder.
type FlunderSpec struct {
	// A name of another flunder, mutually exclusive to the FischerReference.
	FlunderReference string `json:"flunderReference,omitempty" protobuf:"bytes,1,opt,name=flunderReference"`
	// A name of a fischer, mutually exclusive to the FlunderReference.
	FischerReference string `json:"fischerReference,omitempty" protobuf:"bytes,2,opt,name=fischerReference"`
	// The reference type.
	ReferenceType ReferenceType `json:"referenceType,omitempty" protobuf:"bytes,3,opt,name=referenceType"`
}

// FlunderStatus is the status of a Flunder.
type FlunderStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Flunder defines the schema for the "flunders" resource.
type Flunder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   FlunderSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status FlunderStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// GetGroupVersionResource returns a GroupVersionResource with "flunders" as the resource.
// GetGroupVersionResource implements resource.Object
func (Flunder) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource("flunders")
}

// GetObjectMeta implements resource.Object
func (f *Flunder) GetObjectMeta() *metav1.ObjectMeta {
	return &f.ObjectMeta
}

// IsStorageVersion returns true -- v1alpha1.Flunder is used as the internal version.
// IsStorageVersion implements resource.Object.
func (Flunder) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true to indicate Flunder is a namespaced resource.
// NamespaceScoped implements resource.Object.
func (Flunder) NamespaceScoped() bool {
	return true
}

// New implements resource.Object
func (Flunder) New() runtime.Object {
	return &Flunder{}
}

// NewList implements resource.Object
func (Flunder) NewList() runtime.Object {
	return &FlunderList{}
}

// Validate implements resource.Validater
func (f *Flunder) Validate(_ context.Context) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateFlunderSpec(&f.Spec, field.NewPath("spec"))...)
	return allErrs
}

// ValidateUpdate implements resource.ValidateUpdater
func (f *Flunder) ValidateUpdate(ctx context.Context, _ runtime.Object) field.ErrorList {
	return f.Validate(ctx)
}

// validateFlunderSpec validates a FlunderSpec.
func validateFlunderSpec(s *FlunderSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch {
	case len(s.FlunderReference) != 0 && len(s.FischerReference) != 0:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("fischerReference"), s.FischerReference, "cannot be set with flunderReference at the same time"))
	case len(s.FlunderReference) != 0 && s.ReferenceType != FlunderReferenceType:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("flunderReference"), s.FlunderReference, "cannot be set if referenceType is not Flunder"))
	case len(s.FischerReference) != 0 && s.ReferenceType != FischerReferenceType:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("fischerReference"), s.FischerReference, "cannot be set if referenceType is not Fischer"))
	case len(s.FischerReference) == 0 && s.ReferenceType == FischerReferenceType:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("fischerReference"), s.FischerReference, "cannot be empty if referenceType is Fischer"))
	case len(s.FlunderReference) == 0 && s.ReferenceType == FlunderReferenceType:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("flunderReference"), s.FlunderReference, "cannot be empty if referenceType is Flunder"))
	}

	if len(s.ReferenceType) != 0 && s.ReferenceType != FischerReferenceType && s.ReferenceType != FlunderReferenceType {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("referenceType"), s.ReferenceType, "must be Flunder or Fischer"))
	}

	return allErrs
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlunderList is a list of Flunder objects.
type FlunderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Flunder `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// GetListMeta returns the ListMeta
func (c *FlunderList) GetListMeta() *metav1.ListMeta {
	return &c.ListMeta
}
