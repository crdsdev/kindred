# kindred

*NOTE: `kindred` **should not** be used in production environments*

`kindred` is a tool for configuring multiple
[kube-apiserver](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/)
and
[kube-controller-manager](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/)
instances within a single Kubernetes cluster. It bootstraps new instances,
manages access to them, and assists in running [custom
controllers](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#custom-controllers)
against them.

**STATUS**

`kindred` is currently under construction and is only able to bootstrap tenant
Kubernetes instances in [`kind`
v0.7.0](https://github.com/kubernetes-sigs/kind/releases/tag/v0.7.0) clusters.
Accessing tenant instances directly is only available for Linux machines, but
can be configured for other operating systems if [additional
port-mappings](https://kind.sigs.k8s.io/docs/user/configuration/#extra-port-mappings)
are defined.

## Usage

Create tenant Kubernetes API Server and Controller Manager:
```
kindred tenant create
```

List tenant Kubernetes instances:
```
kindred tenant list
```

Get [kubeconfig] for a tenant instance:
```
kindred tenant config <instance-identifier>
```
