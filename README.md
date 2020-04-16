# kindred

*NOTE: `kindred` **should not** be used in production environments at this time*

`kindred` is a tool for configuring multiple
[kube-apiserver](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/)
and
[kube-controller-manager](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/)
instances within a single Kubernetes cluster. It bootstraps new instances,
manages access to them, and assists in running [custom
controllers](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#custom-controllers)
against them.