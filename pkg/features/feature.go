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

// Package features defines feature-gates.
package features

import (
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	"k8s.io/klog/v2"
)

const (
// Every feature gate should add method here following this template:
//
// // owner: @username
// // alpha: v1.X
// MyFeature featuregate.Feature = "MyFeature"
)

func init() {
	if err := utilfeature.DefaultMutableFeatureGate.Add(DefaultKubeFedFeatureGates); err != nil {
		klog.Fatalf("Unexpected error: %v", err)
	}
}

// DefaultKubeFedFeatureGates consists of all known KubeFed-specific
// feature keys.  To add a new feature, define a key for it above and
// add it here.
var DefaultKubeFedFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{}
