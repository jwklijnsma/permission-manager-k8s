# Install

Permission Manager runs as a single Deployment in its own namespace. It needs
Kubernetes **1.24 or newer** (see [compatibility](../reference/compatibility.md)).

You need three values for your cluster:

| Value | Used for | Where to find it |
|---|---|---|
| Cluster name | Name of the cluster/context in the generated kubeconfigs | `kubectl config view --minify -o jsonpath='{.clusters[0].name}'` |
| Control plane address | API server URL written in the generated kubeconfigs, as your users reach it | `kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'` |
| Basic auth password | Password of the `admin` user of the UI | choose one, or [use an auth proxy](../configuration/authentication.md) |

## Install with Helm

The chart is published as an OCI artifact:

```shell
helm upgrade --install permission-manager \
  oci://ghcr.io/jwklijnsma/charts/permission-manager \
  --namespace permission-manager --create-namespace \
  --set config.clusterName=my-cluster \
  --set config.controlPlaneAddress=https://my-cluster.example.com:6443 \
  --set config.basicAuthPassword=changeMe
```

Or with a values file:

```yaml title="my-values.yaml"
config:
  clusterName: my-cluster
  controlPlaneAddress: https://my-cluster.example.com:6443
  basicAuthPassword: changeMe

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: permission-manager.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: permission-manager-tls
      hosts:
        - permission-manager.example.com
```

```shell
helm upgrade --install permission-manager \
  oci://ghcr.io/jwklijnsma/charts/permission-manager \
  --namespace permission-manager --create-namespace --values my-values.yaml
```

!!! tip "Keep the password out of the Helm release"
    Use [`existingSecret`](../configuration/secrets.md) to reference a Secret
    you manage yourself (Sealed Secrets, External Secrets, SOPS...).

All chart values: [Helm values](../reference/helm-values.md).

## Install with kubectl

Create the namespace and the Secret with the configuration:

```shell
kubectl create namespace permission-manager
```

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: permission-manager
  namespace: permission-manager
type: Opaque
stringData:
  PORT: "4000" # port where server is exposed
  CLUSTER_NAME: "my-cluster" # name of the cluster to use in the generated kubeconfig file
  CONTROL_PLANE_ADDRESS: "https://172.17.0.3:6443" # full address of the control plane to use in the generated kubeconfig file
  BASIC_AUTH_PASSWORD: "changeMe" # password used by basic auth (username is `admin`)
```

Then apply the CRD, the default templates and the Deployment from this
repository (`deployments/kubernetes/`):

```shell
kubectl apply -f deployments/kubernetes/seeds/crd.yml
kubectl apply -f deployments/kubernetes/seeds/seed.yml
kubectl apply -f deployments/kubernetes/deploy.yml
```

`deploy.yml` uses the image `ghcr.io/jwklijnsma/permission-manager-k8s:v1.10.0`.

## Open the UI

```shell
# Helm install (service port 80)
kubectl port-forward svc/permission-manager 4000:80 --namespace permission-manager
# kubectl install (service port 4000)
kubectl port-forward svc/permission-manager 4000:4000 --namespace permission-manager
```

The application is now reachable on <http://localhost:4000> (user `admin`).
Continue with the [quick start](quickstart.md).
