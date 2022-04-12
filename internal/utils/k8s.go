package utils

import (
	"fmt"

	extension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	extfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	discfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	dfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sConfig struct {
	Client          kubernetes.Interface
	DynamicClient   dynamic.Interface
	ExtensionClient extension.Interface
	RESTMapper      *restmapper.DeferredDiscoveryRESTMapper
}

func NewFakeK8sConfig() (*K8sConfig, *testing.Fake, *dfake.FakeDynamicClient) {
	client := fake.NewSimpleClientset()
	dclient := dfake.NewSimpleDynamicClient(runtime.NewScheme())
	extclient := extfake.NewSimpleClientset()

	tfake := &testing.Fake{}
	fakeDisc := &discfake.FakeDiscovery{
		Fake: tfake,
		FakedServerVersion: &version.Info{
			GitCommit: "v1.0.0",
		},
	}

	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(fakeDisc))

	return &K8sConfig{
		Client:          client,
		DynamicClient:   dclient,
		ExtensionClient: extclient,
		RESTMapper:      rm,
	}, tfake, dclient
}

// DefaultConfig returns the kubernetes config and client from default configuration.
// The default context is used if context is empty.
func DefaultConfig(context string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if context != "" {
		overrides.CurrentContext = context
	}
	ccfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	config, err := ccfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %w", context, err)
	}
	return config, nil
}

func NewK8sConfig(config *rest.Config) (*K8sConfig, error) {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create dynamic client: %w", err)
	}
	extclient, err := extension.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create extension client: %w", err)
	}

	disc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery client: %w", err)
	}
	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disc))

	return &K8sConfig{
		Client:          client,
		DynamicClient:   dclient,
		ExtensionClient: extclient,
		RESTMapper:      rm,
	}, nil
}

type CommonMetaOptions struct {
	metav1.TypeMeta
	DryRun []string

	ResourceVersion string

	Force        *bool
	FieldManager string

	GracePeriodSeconds *int64
	Preconditions      *metav1.Preconditions
	PropagationPolicy  *metav1.DeletionPropagation
}

func (cmo CommonMetaOptions) GetOptions() metav1.GetOptions {
	return metav1.GetOptions{
		TypeMeta:        cmo.TypeMeta,
		ResourceVersion: cmo.ResourceVersion,
	}
}

func (cmo CommonMetaOptions) PatchOptions() metav1.PatchOptions {
	return metav1.PatchOptions{
		TypeMeta:     cmo.TypeMeta,
		DryRun:       cmo.DryRun,
		Force:        cmo.Force,
		FieldManager: cmo.FieldManager,
	}
}

func (cmo CommonMetaOptions) DeleteOptions() metav1.DeleteOptions {
	return metav1.DeleteOptions{
		TypeMeta:           cmo.TypeMeta,
		GracePeriodSeconds: cmo.GracePeriodSeconds,
		Preconditions:      cmo.Preconditions,
		PropagationPolicy:  cmo.PropagationPolicy,
		DryRun:             cmo.DryRun,
	}
}

func (cmo CommonMetaOptions) CreateOptions() metav1.CreateOptions {
	return metav1.CreateOptions{
		TypeMeta:     cmo.TypeMeta,
		DryRun:       cmo.DryRun,
		FieldManager: cmo.FieldManager,
	}
}
