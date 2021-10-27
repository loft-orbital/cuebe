/*
Copyright © 2021 Loft Orbital

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
package kubernetes

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

var metaAccessor = meta.NewAccessor()

// Deploy deploys obj to the given cluster
// TODO manage obj validation
// TODO returns what has been done (creation, update)
func Deploy(client kubernetes.Interface, restConfig rest.Config, obj runtime.Object) error {
	groupResources, err := restmapper.GetAPIGroupResources(client.Discovery())
	if err != nil {
		return err
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	// Get some metadata needed to make the REST request.
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return err
	}

	// Create a client specifically for creating the object.
	restClient, err := newRestClient(restConfig, mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return err
	}

	// Use the REST helper to create the object in the "default" namespace.
	helper := resource.NewHelper(restClient, mapping)
	name, err := metaAccessor.Name(obj)
	if err != nil {
		return err
	}
	ns, err := metaAccessor.Namespace(obj)
	if err != nil {
		return err
	}
	if ns == "" && helper.NamespaceScoped {
		ns = "default"
	}

	_, err = helper.Get(ns, name)
	if err == nil {
		_, err = helper.Replace(ns, name, true, obj) // Replace if already exist
	} else if apierrors.IsNotFound(err) {
		_, err = helper.Create(ns, true, obj) // Create otherwise
	}

	return err
}
