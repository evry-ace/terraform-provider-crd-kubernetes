Using provider in your project

```
resource "kubernetes_crd" "crd-istio-instance" {
  provider = "kubernetes_crd"

  name        = "istio"
  namespace   = "istio-system"
  api_version = "istio.banzaicloud.io/v1beta1"
  kind = "Istio"
  spec = <<EOF
version: 1.1.5
mtls: true
autoInjectionNamespaces: []
sidecarInjector:
  enabled: true
  initCNIConfiguration:
    enabled: true
EOF
}```
