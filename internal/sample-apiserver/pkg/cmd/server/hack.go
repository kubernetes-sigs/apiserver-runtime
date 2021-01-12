/*
Copyright 2016 The Kubernetes Authors.

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

package server

import (
	"k8s.io/apiserver/pkg/server/options"
)

// ValidateRecommendedOptions validates the options.
//
// A temporary work-around for https://github.com/kubernetes/kubernetes/pull/97954
func ValidateRecommendedOptions(o *options.RecommendedOptions) []error {
	errors := []error{}

	errors = append(errors, o.Etcd.Validate()...)
	errors = append(errors, o.SecureServing.Validate()...)

	if o.Authentication != nil {
		errors = append(errors, o.Authentication.Validate()...)
	}
	if o.Authorization != nil {
		errors = append(errors, o.Authorization.Validate()...)
	}

	errors = append(errors, o.Audit.Validate()...)
	errors = append(errors, o.Features.Validate()...)
	errors = append(errors, o.CoreAPI.Validate()...)
	errors = append(errors, o.Admission.Validate()...)
	errors = append(errors, o.EgressSelector.Validate()...)

	return errors

}
