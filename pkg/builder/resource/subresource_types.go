/*
Copyright 2020 The Kubernetes Authors.

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

package resource

import "net/url"

// StatusSubResource defines interface for registering status subresource to a parent resource.
type StatusSubResource interface {
	SubResource
	// CopyTo copies the content of the status subresource to a parent resource.
	CopyTo(parent ObjectWithStatusSubResource)
}

// SubResource defines interface for registering arbitrary subresource to the parent resource.
type SubResource interface {
	SubResourceName() string
	// TODO: fill the details for this interface.
}

// QueryParameterObject allows the object to be casted to url.Values.
// It's specifically for Connector subresource.
type QueryParameterObject interface {
	ConvertFromUrlValues(values *url.Values) error
}
