package release

#Release: {
  name: string

	// Manifests basic structure
	manifests: [Namespace=string]: [Kind=string]: [Name=string]: {
		kind: Kind
		metadata: name:   Name
		...
	}

	// Only set namespace when Namespace != "$root"
	manifests: [Namespace = (!="$root")]: [Kind=string]: [Name=string]: {
		metadata: namespace: Namespace
	}
}
