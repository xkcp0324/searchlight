/*
Copyright 2017 The Searchlight Authors.

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

// This file was automatically generated by lister-gen

package monitoring

import (
	monitoring "github.com/appscode/searchlight/apis/monitoring"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterAlertLister helps list ClusterAlerts.
type ClusterAlertLister interface {
	// List lists all ClusterAlerts in the indexer.
	List(selector labels.Selector) (ret []*monitoring.ClusterAlert, err error)
	// ClusterAlerts returns an object that can list and get ClusterAlerts.
	ClusterAlerts(namespace string) ClusterAlertNamespaceLister
	ClusterAlertListerExpansion
}

// clusterAlertLister implements the ClusterAlertLister interface.
type clusterAlertLister struct {
	indexer cache.Indexer
}

// NewClusterAlertLister returns a new ClusterAlertLister.
func NewClusterAlertLister(indexer cache.Indexer) ClusterAlertLister {
	return &clusterAlertLister{indexer: indexer}
}

// List lists all ClusterAlerts in the indexer.
func (s *clusterAlertLister) List(selector labels.Selector) (ret []*monitoring.ClusterAlert, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*monitoring.ClusterAlert))
	})
	return ret, err
}

// ClusterAlerts returns an object that can list and get ClusterAlerts.
func (s *clusterAlertLister) ClusterAlerts(namespace string) ClusterAlertNamespaceLister {
	return clusterAlertNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ClusterAlertNamespaceLister helps list and get ClusterAlerts.
type ClusterAlertNamespaceLister interface {
	// List lists all ClusterAlerts in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*monitoring.ClusterAlert, err error)
	// Get retrieves the ClusterAlert from the indexer for a given namespace and name.
	Get(name string) (*monitoring.ClusterAlert, error)
	ClusterAlertNamespaceListerExpansion
}

// clusterAlertNamespaceLister implements the ClusterAlertNamespaceLister
// interface.
type clusterAlertNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ClusterAlerts in the indexer for a given namespace.
func (s clusterAlertNamespaceLister) List(selector labels.Selector) (ret []*monitoring.ClusterAlert, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*monitoring.ClusterAlert))
	})
	return ret, err
}

// Get retrieves the ClusterAlert from the indexer for a given namespace and name.
func (s clusterAlertNamespaceLister) Get(name string) (*monitoring.ClusterAlert, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(monitoring.Resource("clusteralert"), name)
	}
	return obj.(*monitoring.ClusterAlert), nil
}
