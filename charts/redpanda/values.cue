package redpanda

#Values: {
	nameOverride: string | *""
	fullnameOverride: string | *""
	clusterDomain: string | *"cluster.local"

	image: {
	  repository: string | *"docker.redpanda.com/redpandadata/redpanda"
	  tag: string | *""
	  pullPolicy: "IfNotPresent" // TODO add more options
	}

	rbac: {enabled: bool | *true}
	serviceAccount: {
		annotations: {...} // TODO this is to loose
	}
}

#Release: {
	Name: string | *"redpanda"
}

#Context: {Values: #Values, Release: #Release}

in: #Context
