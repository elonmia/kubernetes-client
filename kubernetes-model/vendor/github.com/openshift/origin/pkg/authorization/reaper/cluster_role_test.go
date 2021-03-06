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
package reaper

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgotesting "k8s.io/client-go/testing"
	kapi "k8s.io/kubernetes/pkg/apis/core"

	authorizationapi "github.com/openshift/origin/pkg/authorization/apis/authorization"
	_ "github.com/openshift/origin/pkg/authorization/apis/authorization/install"
	authorizationfake "github.com/openshift/origin/pkg/authorization/generated/internalclientset/fake"
)

func TestClusterRoleReaper(t *testing.T) {
	tests := []struct {
		name                string
		role                *authorizationapi.ClusterRole
		bindings            []*authorizationapi.ClusterRoleBinding
		deletedBindingNames []string
	}{
		{
			name: "no bindings",
			role: &authorizationapi.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "role",
				},
			},
		},
		{
			name: "bindings",
			role: &authorizationapi.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "role",
				},
			},
			bindings: []*authorizationapi.ClusterRoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "binding-1",
					},
					RoleRef: kapi.ObjectReference{Name: "role"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "binding-2",
					},
					RoleRef: kapi.ObjectReference{Name: "role2"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "binding-3",
					},
					RoleRef: kapi.ObjectReference{Name: "role"},
				},
			},
			deletedBindingNames: []string{"binding-1", "binding-3"},
		},
	}

	for _, test := range tests {
		startingObjects := []runtime.Object{}
		startingObjects = append(startingObjects, test.role)
		for _, binding := range test.bindings {
			startingObjects = append(startingObjects, binding)
		}
		tc := authorizationfake.NewSimpleClientset(startingObjects...)

		actualDeletedBindingNames := []string{}
		tc.PrependReactor("delete", "clusterrolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			actualDeletedBindingNames = append(actualDeletedBindingNames, action.(clientgotesting.DeleteAction).GetName())
			return true, nil, nil
		})

		reaper := NewClusterRoleReaper(tc.Authorization(), tc.Authorization(), tc.Authorization())
		err := reaper.Stop("", test.role.Name, 0, nil)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
		}

		expected := sets.NewString(test.deletedBindingNames...)
		actuals := sets.NewString(actualDeletedBindingNames...)
		if !reflect.DeepEqual(expected.List(), actuals.List()) {
			t.Errorf("%s: expected %v, got %v", test.name, expected.List(), actuals.List())
		}
	}
}

func TestClusterRoleReaperAgainstNamespacedBindings(t *testing.T) {
	tests := []struct {
		name                string
		role                *authorizationapi.ClusterRole
		bindings            []*authorizationapi.RoleBinding
		deletedBindingNames []string
	}{
		{
			name: "bindings",
			role: &authorizationapi.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "role",
				},
			},
			bindings: []*authorizationapi.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "binding-1",
						Namespace: "ns-one",
					},
					RoleRef: kapi.ObjectReference{Name: "role"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "binding-2",
						Namespace: "ns-one",
					},
					RoleRef: kapi.ObjectReference{Name: "role2"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "binding-3",
						Namespace: "ns-one",
					},
					RoleRef: kapi.ObjectReference{Name: "role"},
				},
			},
			deletedBindingNames: []string{"binding-1", "binding-3"},
		},
	}

	for _, test := range tests {
		startingObjects := []runtime.Object{}
		startingObjects = append(startingObjects, test.role)
		for _, binding := range test.bindings {
			startingObjects = append(startingObjects, binding)
		}
		tc := authorizationfake.NewSimpleClientset(startingObjects...)

		actualDeletedBindingNames := []string{}
		tc.PrependReactor("delete", "rolebindings", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			actualDeletedBindingNames = append(actualDeletedBindingNames, action.(clientgotesting.DeleteAction).GetName())
			return true, nil, nil
		})

		reaper := NewClusterRoleReaper(tc.Authorization(), tc.Authorization(), tc.Authorization())
		err := reaper.Stop("", test.role.Name, 0, nil)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
		}

		expected := sets.NewString(test.deletedBindingNames...)
		actuals := sets.NewString(actualDeletedBindingNames...)
		if !reflect.DeepEqual(expected.List(), actuals.List()) {
			t.Errorf("%s: expected %v, got %v", test.name, expected.List(), actuals.List())
		}
	}
}
