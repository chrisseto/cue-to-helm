package redpanda

// import "list"

// import corev1 "k8s.io/api/core/v1"

// import rbacv1 "k8s.io/api/rbac/v1"

_cr?: #ClusterRole @template("toYaml")

if in.Values.rbac.enabled {
	_cr: {
		metadata: {
			name:        (#_redpanda.#fullname & in).out
			labels:      (#_full.#labels & in).out
			annotations: in.Values.serviceAccount.annotations
		}
		rules: [{
			apiGroups: [""]
			resources: ["nodes"]
			verbs: ["get", "list"]
		}]
	}
}

out: [for x in [_cr] if x != null { x }]

// apiVersion: rbac.authorization.k8s.io/v1
// kind: ClusterRole
// metadata:
//   name: {{ include "redpanda.fullname" . }}-rpk-bundle
//   labels:
// {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
// {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//     {{- toYaml . | nindent 4 }}
//   {{- end }}
// rules:
//   - apiGroups:
//     - ""
//     resources:
//       - configmaps
//       - endpoints
//       - events
//       - limitranges
//       - persistentvolumeclaims
//       - pods
//       - pods/log
//       - replicationcontrollers
//       - resourcequotas
//       - serviceaccounts
//       - services
//     verbs:
//       - get
//       - list
// ---
// apiVersion: rbac.authorization.k8s.io/v1
// kind: ClusterRoleBinding
// metadata:
//   name: {{ include "redpanda.fullname" . }}
//   labels:
// {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
// {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//     {{- toYaml . | nindent 4 }}
//   {{- end }}
// roleRef:
//   apiGroup: rbac.authorization.k8s.io
//   kind: ClusterRole
//   name: {{ include "redpanda.fullname" . }}
// subjects:
//   - kind: ServiceAccount
//     name: {{ include "redpanda.serviceAccountName" . }}
//     namespace: {{ .Release.Namespace | quote }}
// ---
// apiVersion: rbac.authorization.k8s.io/v1
// kind: ClusterRoleBinding
// metadata:
//   name: {{ include "redpanda.fullname" . }}-rpk-bundle
//   labels:
// {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
// {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//     {{- toYaml . | nindent 4 }}
//   {{- end }}
// roleRef:
//   apiGroup: rbac.authorization.k8s.io
//   kind: ClusterRole
//   name: {{ include "redpanda.fullname" . }}-rpk-bundle
// subjects:
//   - kind: ServiceAccount
//     name: {{ include "redpanda.serviceAccountName" . }}
//     namespace: {{ .Release.Namespace | quote }}
// {{- end }}
// {{- if and .Values.statefulset.sideCars.controllers.enabled .Values.statefulset.sideCars.controllers.createRBAC }}
// ---
// apiVersion: rbac.authorization.k8s.io/v1
// kind: ClusterRole
// metadata:
//   name: {{ include "redpanda.fullname" . }}-sidecar-controllers
//   labels:
//   {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
//   {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//   {{- toYaml . | nindent 4 }}
//   {{- end }}
// rules:
//   - apiGroups:
//       - ""
//     resources:
//       - nodes
//     verbs:
//       - get
//       - list
//       - watch
//   - apiGroups:
//       - ""
//     resources:
//       - persistentvolumes
//     verbs:
//       - delete
//       - get
//       - list
//       - patch
//       - update
//       - watch
// ---
// apiVersion: rbac.authorization.k8s.io/v1
// kind: ClusterRoleBinding
// metadata:
//   name: {{ include "redpanda.fullname" . }}-sidecar-controllers
//   labels:
//   {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
//   {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//   {{- toYaml . | nindent 4 }}
//     {{- end }}
// roleRef:
//   apiGroup: rbac.authorization.k8s.io
//   kind: ClusterRole
//   name: {{ include "redpanda.fullname" . }}-sidecar-controllers
// subjects:
//   - kind: ServiceAccount
//     name: {{ include "redpanda.serviceAccountName" . }}
//     namespace: {{ .Release.Namespace | quote }}
// ---
// apiVersion: rbac.authorization.k8s.io/v1
// kind: Role
// metadata:
//   name: {{ include "redpanda.fullname" . }}-sidecar-controllers
//   labels:
//   {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
//   {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//   {{- toYaml . | nindent 4 }}
//     {{- end }}
// rules:
//   - apiGroups:
//       - apps
//     resources:
//       - statefulsets/status
//     verbs:
//       - patch
//       - update
//   - apiGroups:
//       - ""
//     resources:
//       - secrets
//       - pods
//     verbs:
//       - get
//       - list
//       - watch
//   - apiGroups:
//       - apps
//     resources:
//       - statefulsets
//     verbs:
//       - get
//       - patch
//       - update
//       - list
//       - watch
//   - apiGroups:
//       - ""
//     resources:
//       - persistentvolumeclaims
//     verbs:
//       - delete
//       - get
//       - list
//       - patch
//       - update
//       - watch
// ---
// apiVersion: rbac.authorization.k8s.io/v1
// kind: RoleBinding
// metadata:
//   name: {{ include "redpanda.fullname" . }}-sidecar-controllers
//   labels:
//   {{- with include "full.labels" . }}
//   {{- . | nindent 4 }}
//   {{- end }}
//   {{- with .Values.serviceAccount.annotations }}
//   annotations:
//   {{- toYaml . | nindent 4 }}
//   {{- end }}
// roleRef:
//   apiGroup: rbac.authorization.k8s.io
//   kind: Role
//   name: {{ include "redpanda.fullname" . }}-sidecar-controllers
// subjects:
//   - kind: ServiceAccount
//     name: {{ include "redpanda.serviceAccountName" . }}
//     namespace: {{ .Release.Namespace | quote }}
// {{- end }}
