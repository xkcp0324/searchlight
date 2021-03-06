/*
Copyright 2019 The Searchlight Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/appscode/searchlight/apis/monitoring/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeIncidents implements IncidentInterface
type FakeIncidents struct {
	Fake *FakeMonitoringV1alpha1
	ns   string
}

var incidentsResource = schema.GroupVersionResource{Group: "monitoring.appscode.com", Version: "v1alpha1", Resource: "incidents"}

var incidentsKind = schema.GroupVersionKind{Group: "monitoring.appscode.com", Version: "v1alpha1", Kind: "Incident"}

// Get takes name of the incident, and returns the corresponding incident object, and an error if there is any.
func (c *FakeIncidents) Get(name string, options v1.GetOptions) (result *v1alpha1.Incident, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(incidentsResource, c.ns, name), &v1alpha1.Incident{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Incident), err
}

// List takes label and field selectors, and returns the list of Incidents that match those selectors.
func (c *FakeIncidents) List(opts v1.ListOptions) (result *v1alpha1.IncidentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(incidentsResource, incidentsKind, c.ns, opts), &v1alpha1.IncidentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.IncidentList{ListMeta: obj.(*v1alpha1.IncidentList).ListMeta}
	for _, item := range obj.(*v1alpha1.IncidentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested incidents.
func (c *FakeIncidents) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(incidentsResource, c.ns, opts))

}

// Create takes the representation of a incident and creates it.  Returns the server's representation of the incident, and an error, if there is any.
func (c *FakeIncidents) Create(incident *v1alpha1.Incident) (result *v1alpha1.Incident, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(incidentsResource, c.ns, incident), &v1alpha1.Incident{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Incident), err
}

// Update takes the representation of a incident and updates it. Returns the server's representation of the incident, and an error, if there is any.
func (c *FakeIncidents) Update(incident *v1alpha1.Incident) (result *v1alpha1.Incident, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(incidentsResource, c.ns, incident), &v1alpha1.Incident{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Incident), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeIncidents) UpdateStatus(incident *v1alpha1.Incident) (*v1alpha1.Incident, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(incidentsResource, "status", c.ns, incident), &v1alpha1.Incident{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Incident), err
}

// Delete takes name of the incident and deletes it. Returns an error if one occurs.
func (c *FakeIncidents) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(incidentsResource, c.ns, name), &v1alpha1.Incident{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIncidents) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(incidentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.IncidentList{})
	return err
}

// Patch applies the patch and returns the patched incident.
func (c *FakeIncidents) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Incident, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(incidentsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Incident{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Incident), err
}
