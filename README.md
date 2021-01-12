# Kubebuilder: apiserver-runtime

**Note**: This project is Alpha with no commitments from its maintainers to provide any sort of support.
If you intend to use this project in production, we recommend you become involved in maintaining it.

## Experimental / Work in Progress

This project exists to explore the creation of new libraries for building Kubernetes API extension servers.

This project is experimental and **may change drastically or be cancelled with little or no public notice**.

If deletion of this repo would negatively impact you, then make sure this is documented:

- Comment on this [issue](https://github.com/kubernetes-sigs/apiserver-runtime/issues/7)

And make sure you will get any public notifications about this project.

- Join the kubebuilder@googlegroups.com mailing list

Documenting your dependence on this project does not guarantee that it will not be cancelled if there was insufficient
interest on the part of the maintainers to continue maintaining it.

## Goals

apiserver-runtime provides libraries for building on top of the Kubernetes apiserver and apimachinery modules.

If you are simply looking to create new Kubernetes extension resource types backed by etcd storage, then you should
use CRDs instead of this library.  This library is only for the cases where you want to expose Kubernetes endpoints
which are not backed by etcd storage.

apiserver-runtime is developed under the [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) subproject.

See the docs: [https://pkg.go.dev/sigs.k8s.io/apiserver-runtime/pkg/builder](https://pkg.go.dev/sigs.k8s.io/apiserver-runtime/pkg/builder)

See the example: [sample/cmd/apiserver/main.go](sample/cmd/apiserver/main.go)

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/forum/#!forum/kubebuilder)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE
