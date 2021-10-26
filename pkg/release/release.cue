package release

#Release: {
	name: string

	manifests: {
		[string]: [Kind=string]: [Name=string]: {
			kind: Kind
			metadata: name: Name
			...
		}

		[Ns = !="$global"]: [string]: [string]: {
			metadata: namespace: Ns
		}
	}
}
