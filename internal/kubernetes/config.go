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

	"cuelang.org/go/cue"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Import every auth client
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

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

func ExtractContext(v cue.Value, path string) (string, error) {
	p := cue.ParsePath(path)
	if p.Err() != nil {
		return "", fmt.Errorf("failed to extract kubernetes context: %w", p.Err())
	}
	ktx := v.LookupPath(p)
	if ktx.Err() != nil {
		return "", fmt.Errorf("failed to extract kubernetes context: %w", ktx.Err())
	}
	sktx, err := ktx.String()
	if err != nil {
		return "", fmt.Errorf("failed to extract kubernetes context: %w", err)
	}
	return sktx, nil
}
