package release

#Release: {
	sauces: [string]: string

	manifests: [Kind=string]: [Name=string]: {
		kind: Kind
		metadata: name: Name
	}

	// Namespace
	manifests: Namespace: potato: apiVersion: "v1"

	// ConfigMap
	manifests: ConfigMap: sauce: {
		apiVersion: "v1"
		metadata: namespace: manifests.Namespace.potato.metadata.name
		data: sauces
	}
}
