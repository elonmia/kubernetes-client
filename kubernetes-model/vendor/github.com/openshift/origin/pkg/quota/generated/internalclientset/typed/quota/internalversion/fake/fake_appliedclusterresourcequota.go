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
package fake

import (
	quota "github.com/openshift/origin/pkg/quota/apis/quota"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeAppliedClusterResourceQuotas implements AppliedClusterResourceQuotaInterface
type FakeAppliedClusterResourceQuotas struct {
	Fake *FakeQuota
	ns   string
}

var appliedclusterresourcequotasResource = schema.GroupVersionResource{Group: "quota.openshift.io", Version: "", Resource: "appliedclusterresourcequotas"}

var appliedclusterresourcequotasKind = schema.GroupVersionKind{Group: "quota.openshift.io", Version: "", Kind: "AppliedClusterResourceQuota"}

// Get takes name of the appliedClusterResourceQuota, and returns the corresponding appliedClusterResourceQuota object, and an error if there is any.
func (c *FakeAppliedClusterResourceQuotas) Get(name string, options v1.GetOptions) (result *quota.AppliedClusterResourceQuota, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(appliedclusterresourcequotasResource, c.ns, name), &quota.AppliedClusterResourceQuota{})

	if obj == nil {
		return nil, err
	}
	return obj.(*quota.AppliedClusterResourceQuota), err
}

// List takes label and field selectors, and returns the list of AppliedClusterResourceQuotas that match those selectors.
func (c *FakeAppliedClusterResourceQuotas) List(opts v1.ListOptions) (result *quota.AppliedClusterResourceQuotaList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(appliedclusterresourcequotasResource, appliedclusterresourcequotasKind, c.ns, opts), &quota.AppliedClusterResourceQuotaList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &quota.AppliedClusterResourceQuotaList{}
	for _, item := range obj.(*quota.AppliedClusterResourceQuotaList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}
