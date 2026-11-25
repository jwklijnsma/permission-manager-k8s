# Testing and CI

## Unit tests

=== "Go server"

    ```shell
    make build-ui            # or: mkdir -p static/build && touch static/build/index.html
    go test ./...
    ```

    The server embeds the web-client build (`static/build`), so a build or a
    placeholder `index.html` must exist. Tests use the client-go fake clientset;
    no cluster is needed.

=== "Web-client"

    ```shell
    cd web-client
    yarn install --frozen-lockfile
    CI=true yarn test --watchAll=false
    npx tsc --noEmit -p .
    ```

    On Node 17+ the production build needs `NODE_OPTIONS=--openssl-legacy-provider`
    (react-scripts 3 / webpack 4).

## What the tests cover

| Area | Tests |
|---|---|
| Configuration | required variables, basic auth on/off and invalid values, token mode and expiration, subject kinds (`internal/config`) |
| Kubeconfig | existing Secret, Secret created and filled asynchronously, timeout instead of empty token, stale Secret of a recreated ServiceAccount, non-token Secret, TokenRequest mode and expiration, CA from `kube-root-ca.crt`, user deletion (`internal/resources`) |
| RBAC | legacy binding detection, RoleBinding/ClusterRoleBinding creation and labels (`internal/resources`) |
| HTTP API | basic auth enforced / disabled, subject normalisation, invalid role and subject kinds, validation stops the handler, username validation, 500 on failed kubeconfig (`internal/server`) |
| Web-client | read/write verb mapping, binding requests sent by the user wizard |

## GitHub Actions

| Workflow | When | What |
|---|---|---|
| `ci` | every push and PR | gofmt, `go vet`, `go test -race` with coverage, server build; web-client type check, Jest, build; `helm lint` + `helm template` assertions + kubeconform; shellcheck |
| `docs` | changes under `docs/` | `zensical build --strict`, publish to GitHub Pages from `main` |
| `docker` | PRs touching the server/UI, `main`, `v*` tags | build the image, smoke test, push to GHCR (not on PRs) |
| `helm-release` | `helm-chart-v*` tags | push the chart and the Artifact Hub metadata to `oci://ghcr.io/<owner>/charts` |
| `renovate` | every 4 hours | dependency update PRs (Go modules with `go mod tidy`, npm, Actions, Docker) |

## End-to-end tests

The e2e suite creates a kind cluster, installs the chart and runs API checks and
Cypress tests:

```shell
CLUSTER_VERSION=1.34.0 KIND_VERSION=0.30.0 HELM_VERSION=3.19.0 make test-e2e
```
