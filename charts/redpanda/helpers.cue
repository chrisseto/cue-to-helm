package redpanda

// TODO trunc 63 + trimSuffix
// let redpanda_fullname = "\(in.Release.Name)"
// if in.Values.fullnameOverride {
// 	let redpanda_fullname = in.Values.fullnameOverride
// }

#_redpanda: {
	#fullname: {
		in:  #Context
		out: "\(in.Release.Name)"
		if in.Values.fullnameOverride != null {
			out: in.Values.fullnameOverride
		}
	}
}

#_full: {
	#labels: {
		in: #Context
		out: {}
	}
}
