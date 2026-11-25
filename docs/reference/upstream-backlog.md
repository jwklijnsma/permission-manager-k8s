# Upstream issues and pull requests

Status in this fork of every issue and pull request that was open on
[sighupio/permission-manager](https://github.com/sighupio/permission-manager)
on 2026-11-25.

## Pull requests

| PR | Title | Status in this fork |
|---|---|---|
| [#194](https://github.com/sighupio/permission-manager/pull/194) | Add User/Group subjects to generated RBAC bindings | **Done differently**: server side and configurable (`ADDITIONAL_SUBJECT_KINDS`) instead of a hard-coded web-client constant; the web-client test was adapted. The committed `run-server` binary was left out |
| [#183](https://github.com/sighupio/permission-manager/pull/183) | Fix indents in CRD definition | **Merged** |
| [#182](https://github.com/sighupio/permission-manager/pull/182) | Update validUsernameRegex | **Not merged, test adopted**: removing the `\.` escape makes `.` match any character (`gino_pino`, `a/b` would pass). The regex is kept, its test cases were added and extended |
| [#180](https://github.com/sighupio/permission-manager/pull/180), [#178](https://github.com/sighupio/permission-manager/pull/178) | bump word-wrap 1.2.3 → 1.2.4 | **Merged** |
| [#177](https://github.com/sighupio/permission-manager/pull/177), [#176](https://github.com/sighupio/permission-manager/pull/176) | bump semver 5.7.1 → 5.7.2 | **Merged** |
| [#172](https://github.com/sighupio/permission-manager/pull/172) | Add secret dynamically | **Done differently**: `existingSecret` + `secret.annotations`; the PR checked `secret.exists` while the values defined `secret.exist`, so it never worked |
| [#179](https://github.com/sighupio/permission-manager/pull/179) | Replace Cypress with Playwright | **Not merged**: built on top of the unfinished UI rewrite of #166 |
| [#166](https://github.com/sighupio/permission-manager/pull/166) | Update UI stack | **Not merged**: unfinished rewrite (new views next to the old ones, 5000+ lockfile lines); fixes were applied to the current UI instead |
| [#165](https://github.com/sighupio/permission-manager/pull/165) | Evolve frontend deployment | **Not merged**: removes the embedded UI from the server and adds a separate nginx Deployment, a larger architecture change |

## Issues

| Issue | Title | Status |
|---|---|---|
| [#193](https://github.com/sighupio/permission-manager/issues/193) | K8S v1.34 | **Fixed**: see [compatibility](compatibility.md), kind config fixed for current node images |
| [#192](https://github.com/sighupio/permission-manager/issues/192) | is this project dead? | This fork is maintained |
| [#191](https://github.com/sighupio/permission-manager/issues/191) | Create permission-manager users through API or command | **Documented**: [HTTP API](../usage/api.md) and GitOps examples |
| [#190](https://github.com/sighupio/permission-manager/issues/190) | Make basic auth optional | **Fixed**: `BASIC_AUTH_DISABLED` / `config.basicAuthEnabled`, fail closed, see [authentication](../configuration/authentication.md) |
| [#189](https://github.com/sighupio/permission-manager/issues/189) | Index the Helm chart on ArtifactHub | **Fixed**: OCI chart + `artifacthub-repo.yml` + Artifact Hub annotations, see [releasing](../development/releasing.md) |
| [#186](https://github.com/sighupio/permission-manager/issues/186) | "read" is not a valid verb | **Fixed**: read = `get`/`list`/`watch`, see [templates](../usage/templates.md) |
| [#173](https://github.com/sighupio/permission-manager/issues/173) | Kubernetes version confirmation | **Fixed**: [compatibility](compatibility.md) |
| [#171](https://github.com/sighupio/permission-manager/issues/171) | Use an existing secret in the chart | **Fixed**: `existingSecret` |
| [#170](https://github.com/sighupio/permission-manager/issues/170) | Annotations on the chart Secret | **Fixed**: `secret.annotations` |
| [#169](https://github.com/sighupio/permission-manager/issues/169) | Wrong chart appVersion format | **Fixed**: `appVersion: "1.10.0"`, image tag defaults to `v` + appVersion |
| [#168](https://github.com/sighupio/permission-manager/issues/168) | Kubeconfig generation on Kubernetes > 1.24 | **Fixed**: robust token Secret handling, errors instead of empty tokens, optional TokenRequest, see [kubeconfig tokens](../configuration/kubeconfig-tokens.md) |
| [#157](https://github.com/sighupio/permission-manager/issues/157) | Fix and improve rename-helm-release.sh | **Fixed**: `set -eu`, tool checks, configurable repository, error handling, idempotent, JSON built with `jq` |
| [#138](https://github.com/sighupio/permission-manager/issues/138) | UX: namespace dropdown | Open: UI redesign |
| [#117](https://github.com/sighupio/permission-manager/issues/117) | Select different subject kinds | **Fixed** (server side): [binding subjects](../configuration/subjects.md) |
| [#116](https://github.com/sighupio/permission-manager/issues/116) | Users in config as code | **Documented**: [users as code](../usage/api.md#users-as-code-gitops) |
| [#91](https://github.com/sighupio/permission-manager/issues/91) | Add subjects to (cluster)rolebindings | **Fixed**: [binding subjects](../configuration/subjects.md) |
| [#84](https://github.com/sighupio/permission-manager/issues/84) | Broken access view for custom template | **Fixed**: explicit verbs and rules without `resources` are handled |
| [#68](https://github.com/sighupio/permission-manager/issues/68) | Document supported Kubernetes versions | **Fixed**: [compatibility](compatibility.md) |
| [#67](https://github.com/sighupio/permission-manager/issues/67) | Upgrade process | **Documented**: [upgrade](../getting-started/upgrade.md) |
| [#42](https://github.com/sighupio/permission-manager/issues/42) | Allow email as username | Not possible for ServiceAccount names; use `User` subjects, see [users](../usage/users.md#usernames) |
| [#41](https://github.com/sighupio/permission-manager/issues/41) | Flexible subjects for bindings | **Fixed**: [binding subjects](../configuration/subjects.md) |
| [#38](https://github.com/sighupio/permission-manager/issues/38) | Flexible templates for clusterroles | Partly: any ClusterRole with the template prefix is a template, see [templates](../usage/templates.md) |
| [#19](https://github.com/sighupio/permission-manager/issues/19) | Admin user and normal user | Open: needs a multi-user authorization model |

## Other fixes found on the way

- Validation errors were written to the response but the handler **kept going**
  and still created the object. Requests with missing fields are now rejected
  before anything is created.
- Deleting a user left its ServiceAccount and token behind, so its kubeconfig kept
  working. The ServiceAccount and Secret are now deleted with the user.
- ClusterRoleBindings created from the UI had an empty `generated_for_user` label
  (the web-client sent `generated_for_user`, the server read `user`).
- The username in `delete-user` was concatenated into an API path without
  validation; usernames are now validated on every endpoint that uses them.
- The Kubernetes client was created on every request; it is now created once.
- A previous commit in this fork set `ingress.enabled: true` with a placeholder
  host, `appVersion: 1.0.0`, a meaningless `servicePort` value and an invalid
  kubeadm version in the kind config; these were reverted or corrected.
