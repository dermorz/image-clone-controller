module github.com/dermorz/image-clone-controller

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/google/go-containerregistry v0.4.1
	github.com/google/go-containerregistry/pkg/authn/k8schain v0.0.0-20210316173552-70c58c0e4786
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.3
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	sigs.k8s.io/controller-runtime v0.7.0
)
