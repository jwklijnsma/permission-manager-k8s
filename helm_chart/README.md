# Permission Manager Helm Chart

A Helm chart for Permission Manager, a simple to use web application that enables a super-easy and
user-friendly RBAC management for Kubernetes.

Full documentation: <https://jwklijnsma.github.io/permission-manager-k8s/>

## Install

```bash
helm install permission-manager oci://ghcr.io/jwklijnsma/charts/permission-manager \
  --namespace permission-manager --create-namespace \
  --set config.clusterName=my-cluster \
  --set config.controlPlaneAddress=https://my-cluster.example.com:6443 \
  --set config.basicAuthPassword=changeme
```

Requires Kubernetes >= 1.24.

## Configuration options

<!-- --8<-- [start:configuration] -->
| Parameter | Description | Default |
| --- | --- | --- |
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image | `ghcr.io/jwklijnsma/permission-manager-k8s` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag, defaults to `v` + chart `appVersion` | `""` |
| `imagePullSecrets` | Image pull secrets | `[]` |
| `nameOverride` / `fullnameOverride` | Override resource names | `""` |
| `serviceAccount.create` | Create the ServiceAccount used by Permission Manager | `true` |
| `serviceAccount.annotations` | ServiceAccount annotations | `{}` |
| `serviceAccount.name` | ServiceAccount name | `""` |
| `podAnnotations` | Pod annotations | `{}` |
| `podSecurityContext` | Pod security context | `{fsGroup: 2000}` |
| `securityContext` | Container security context | `{}` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `service.nodePort` | Node port when `service.type=NodePort` | `null` |
| `ingress.enabled` | Create an Ingress | `false` |
| `ingress.className` | IngressClass name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts and paths | see `values.yaml` |
| `ingress.tls` | Ingress TLS configuration | unset |
| `resources` | Container resources | `{}` |
| `autoscaling.*` | HorizontalPodAutoscaler settings | disabled |
| `nodeSelector` / `tolerations` / `affinity` | Scheduling | `{}` / `[]` / `{}` |
| `existingSecret` | Use an existing Secret with `CLUSTER_NAME`, `CONTROL_PLANE_ADDRESS`, `BASIC_AUTH_PASSWORD` instead of creating one | `""` |
| `secret.annotations` | Annotations of the Secret created by the chart | `{}` |
| `config.clusterName` | Cluster name written in the generated kubeconfigs (required unless `existingSecret`) | `""` |
| `config.controlPlaneAddress` | API server address written in the generated kubeconfigs (required unless `existingSecret`) | `""` |
| `config.basicAuthEnabled` | Protect the UI/API with basic auth. Only disable behind an authenticating proxy | `true` |
| `config.basicAuthUsername` | Basic auth username | `admin` |
| `config.basicAuthPassword` | Basic auth password (required when basic auth is enabled, unless `existingSecret`) | `""` |
| `config.additionalSubjectKinds` | Extra subject kinds (`User`, `Group`) added to every generated binding | `[]` |
| `config.kubeconfig.tokenMode` | `secret` (non-expiring token Secret) or `tokenrequest` (time-bound token) | `secret` |
| `config.kubeconfig.tokenExpiration` | Token validity for `tokenrequest` mode (Go duration, min `10m`) | `8760h` |
| `config.templates` | Namespaced ClusterRole templates offered in the UI | `operation`, `developer` |
<!-- --8<-- [end:configuration] -->
