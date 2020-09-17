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

package sample

import (
	"sigs.k8s.io/apiserver-runtime/sample/pkg/apis/sample/v1alpha1"
)

// Fischer is the internal type used for Fischer
type Fischer = v1alpha1.Fischer

// FischerList is the internal type used for FischerList
type FischerList = v1alpha1.FischerList

// Flunder is the internal type used for Flunder
type Flunder = v1alpha1.Flunder

// FlunderList is the internal type used for FlunderList
type FlunderList = v1alpha1.FlunderList
