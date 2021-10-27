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
	"fmt"
	"strings"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const lastAppliedConfig = "cuebe.loftorbital.com/last-applied-configuration"
const defaultNamespace = "default"

var (
	metaAccessor = meta.NewAccessor()
	annotator    = patch.NewAnnotator(lastAppliedConfig)
	patchMaker   = patch.NewPatchMaker(annotator)
)

// Deploy deploys obj to the given cluster.
// If the obj already exist in the cluster, a three way merge patch will be computed.
// If the patch is not empty, the current object will be replaced by obj.
// TODO support dry-run
// TODO manage obj validation
func Deploy(client kubernetes.Interface, restConfig rest.Config, obj runtime.Object) error {
	groupResources, err := restmapper.GetAPIGroupResources(client.Discovery())
	if err != nil {
		return err
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	// Get some metadata needed to make the REST request.
	gvk := obj.GetObjectKind().GroupVersionKind()
	mapping, err := rm.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return err
	}

	// Create a client specifically for creating the object.
	restClient, err := newRestClient(restConfig, mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return err
	}

	// Get useful object meta
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
		ns = defaultNamespace // Set default namespace for namespaced resource with no namespace
	}

	current, err := helper.Get(ns, name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		// Object does not exist yet
		return create(obj, ns, helper)
	}
	// Object already exist, check if we need to update it
	patchResult, err := patchMaker.Calculate(current, obj, patch.IgnoreStatusFields())
	if err != nil {
		return err
	}

	if patchResult.IsEmpty() {
		// Nothing to do
		fmt.Println(buildFeedback(gvk.Kind, name, "unchanged", helper.ServerDryRun))
		return nil
	}

	// Update the object
	return update(obj, ns, name, helper)
}

func update(obj runtime.Object, ns, name string, helper *resource.Helper) error {
	if err := annotator.SetLastAppliedAnnotation(obj); err != nil { // Add last configuration annotation
		return err
	}
	nobj, err := helper.Replace(ns, name, true, obj)
	if err != nil {
		return err
	}
	gvk := nobj.GetObjectKind().GroupVersionKind()
	if err != nil {
		return err
	}
	fmt.Println(buildFeedback(gvk.Kind, name, "updated", helper.ServerDryRun))
	return nil
}

func create(obj runtime.Object, ns string, helper *resource.Helper) error {
	if err := annotator.SetLastAppliedAnnotation(obj); err != nil { // Add last configuration annotation
		return err
	}
	nobj, err := helper.Create(ns, true, obj) // Create the object
	if err != nil {
		return err
	}
	gvk := nobj.GetObjectKind().GroupVersionKind()
	name, err := metaAccessor.Name(obj)
	if err != nil {
		return err
	}
	fmt.Println(buildFeedback(gvk.Kind, name, "created", helper.ServerDryRun))
	return nil
}

func buildFeedback(kind, name, verb string, dryrun bool) string {
	s := fmt.Sprintf("%s/%s %s", kind, name, verb)
	if dryrun {
		s += " (server dry run)"
	}
	return strings.ToLower(s)
}
