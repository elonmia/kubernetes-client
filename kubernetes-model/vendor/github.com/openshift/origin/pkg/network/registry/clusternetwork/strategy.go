/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package clusternetwork

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	networkapi "github.com/openshift/origin/pkg/network/apis/network"
	"github.com/openshift/origin/pkg/network/apis/network/validation"
)

// sdnStrategy implements behavior for ClusterNetworks
type sdnStrategy struct {
	runtime.ObjectTyper
}

// Strategy is the default logic that applies when creating and updating ClusterNetwork
// objects via the REST API.
var Strategy = sdnStrategy{legacyscheme.Scheme}

var _ rest.GarbageCollectionDeleteStrategy = sdnStrategy{}

func (sdnStrategy) DefaultGarbageCollectionPolicy(ctx apirequest.Context) rest.GarbageCollectionPolicy {
	return rest.Unsupported
}

func (sdnStrategy) PrepareForUpdate(ctx apirequest.Context, obj, old runtime.Object) {}

// NamespaceScoped is false for sdns
func (sdnStrategy) NamespaceScoped() bool {
	return false
}

func (sdnStrategy) GenerateName(base string) string {
	return base
}

func (sdnStrategy) PrepareForCreate(ctx apirequest.Context, obj runtime.Object) {
}

// Canonicalize normalizes the object after validation.
func (sdnStrategy) Canonicalize(obj runtime.Object) {
}

// Validate validates a new sdn
func (sdnStrategy) Validate(ctx apirequest.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateClusterNetwork(obj.(*networkapi.ClusterNetwork))
}

// AllowCreateOnUpdate is false for sdn
func (sdnStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (sdnStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for a ClusterNetwork
func (sdnStrategy) ValidateUpdate(ctx apirequest.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateClusterNetworkUpdate(obj.(*networkapi.ClusterNetwork), old.(*networkapi.ClusterNetwork))
}
