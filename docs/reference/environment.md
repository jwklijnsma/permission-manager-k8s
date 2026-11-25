# Environment variables

The server is configured only through environment variables. Invalid values stop
the server at start-up with an error message.

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | yes | - | Port the server listens on (the chart sets `4000`) |
| `NAMESPACE` | yes | - | Namespace of Permission Manager, where user ServiceAccounts live (the chart sets it from the pod namespace) |
| `CLUSTER_NAME` | yes | - | Cluster name in the generated kubeconfigs |
| `CONTROL_PLANE_ADDRESS` | yes | - | API server URL in the generated kubeconfigs |
| `BASIC_AUTH_PASSWORD` | when basic auth is on | - | Basic auth password |
| `BASIC_AUTH_USERNAME` | no | `admin` | Basic auth username |
| `BASIC_AUTH_DISABLED` | no | `false` | `true` disables basic auth, see [authentication](../configuration/authentication.md). Non-boolean values are rejected |
| `ADDITIONAL_SUBJECT_KINDS` | no | empty | Comma separated `User`, `Group`, see [binding subjects](../configuration/subjects.md) |
| `KUBECONFIG_TOKEN_MODE` | no | `secret` | `secret` or `tokenrequest`, see [kubeconfig tokens](../configuration/kubeconfig-tokens.md) |
| `KUBECONFIG_TOKEN_EXPIRATION` | no | `8760h` | Go duration, minimum `10m`, used by `tokenrequest` |

The Kubernetes client uses the in-cluster configuration, or `KUBECONFIG` /
`~/.kube/config` when run outside a cluster (local development).
