ns: {
	apiVersion: "v1"
	kind:       "Namespace"
	metadata: name: "potato"
}

cm: {
	apiVersion: "v1"
	kind:       "ConfigMap"
	metadata: {
		name:      "sauce"
		namespace: ns.metadata.name
	}
	data: tomato: "ketchup"
}
