# Secrets

Permission Manager reads `CLUSTER_NAME`, `CONTROL_PLANE_ADDRESS` and
`BASIC_AUTH_PASSWORD` from a Secret.

## Secret created by the chart

By default the chart creates the Secret from `config.clusterName`,
`config.controlPlaneAddress` and `config.basicAuthPassword`. The values are
stored in the Helm release, and a change of the Secret rolls the Deployment
(`checksum/secret` annotation).

Annotations can be added to it (upstream issue #170), for example for
[Reloader](https://github.com/stakater/Reloader):

```yaml
secret:
  annotations:
    reloader.stakater.com/match: "true"
```

## Existing Secret

To keep the values out of Helm (upstream issue #171, PR #172), create the Secret
yourself and reference it:

```yaml
existingSecret: permission-manager-config
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: permission-manager-config
  namespace: permission-manager
type: Opaque
stringData:
  CLUSTER_NAME: my-cluster
  CONTROL_PLANE_ADDRESS: https://my-cluster.example.com:6443
  BASIC_AUTH_PASSWORD: changeMe   # not needed with config.basicAuthEnabled=false
```

With `existingSecret` set, the chart creates no Secret and ignores
`config.clusterName`, `config.controlPlaneAddress` and `config.basicAuthPassword`.
`PORT` and `NAMESPACE` are set by the Deployment.
