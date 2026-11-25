# OpenCode Remediation Plan — sighupio/permission-manager

> **Goal:** Use OpenCode with a local Ollama model to inspect, test, and safely remediate the repository.
>
> **Repository:** https://github.com/sighupio/permission-manager  
> **Prepared:** 2026-11-25  
> **Mode:** Evidence-driven, incremental changes; no automatic releases or pushes.

## 1. Agent operating instructions

You are a senior Go/Kubernetes engineer working in this repository. Use the local Ollama model through the already-configured OpenCode provider.

### Non-negotiable rules

1. Read the repository's `AGENTS.md`, `CONTRIBUTING*`, `README*`, Makefile, CI workflows, and development documentation before editing.
2. Treat the checked-out source and tests as authoritative. GitHub issue/PR descriptions are leads, not proof that a bug still exists.
3. Start with `git status`, current branch, and recent history. Never overwrite, reset, clean, or reformat unrelated user changes.
4. Do not assume a proposed fix is correct. Reproduce the behavior, trace the code path, add a regression test, then implement the smallest compatible change.
5. Work on one independently testable issue at a time. Keep diffs small and reviewable.
6. Preserve existing public APIs, CRD compatibility, Helm values, labels, annotations, and backwards compatibility unless a task explicitly requires a change.
7. Do not upgrade major dependencies, regenerate large files, or change generated artifacts without understanding the repository's generation process.
8. Never commit, push, open a PR, publish an image/chart, or deploy to a cluster unless the human explicitly asks.
9. Never access production clusters or real credentials. Use local tests, disposable clusters, mocks, and test fixtures only.
10. If evidence is insufficient, tests cannot be run, or a change could break compatibility, stop and report the uncertainty instead of guessing.

### Required workflow for every task

- Locate the relevant implementation, tests, docs, manifests, and history.
- Read the full linked issue/PR and inspect comments/review feedback with `gh` if authenticated, otherwise use the public GitHub pages.
- Determine whether the issue is already fixed, superseded, or addressed by another open PR.
- Write a short hypothesis and reproduction/test plan before editing.
- Add or update a failing regression test first when practical.
- Implement the smallest fix.
- Run focused tests, then relevant package/module tests, lint/build, and integration tests where feasible.
- Inspect `git diff --check` and the complete diff.
- Update documentation and examples when behavior or configuration changes.
- Report exact commands and results. Clearly distinguish passed, failed, skipped, and not-run checks.

## 2. Repository baseline and discovery

First build a factual map of the codebase. Do not begin broad refactoring.

- [ ] Record `git status --short --branch`, `git log -10 --oneline`, Go version, Node/package-manager versions, and tool versions.
- [ ] Identify Go modules, main packages, controllers, API types, CRDs, RBAC generation, frontend, Helm chart(s), deployment manifests, and e2e tests.
- [ ] Map each component to its build/test/lint commands from the Makefile and CI.
- [ ] Identify supported Kubernetes versions from docs, CI matrices, dependencies, release metadata, and actual API usage.
- [ ] Identify generated files and their generators; do not hand-edit generated output.
- [ ] Run the existing test/build/lint suite once before changes and save the baseline.
- [ ] Note pre-existing failures separately so they are not attributed to the patch.

Suggested initial commands (adapt to repository tooling):

```bash
git status --short --branch
git log -10 --oneline
find . -maxdepth 3 -type f | sort | head -300
grep -RInE 'TODO|FIXME|Deprecated|kubernetes.io/service-account-token' --exclude-dir=.git .
make help 2>/dev/null || true
go version
go test ./...
```

Do not run an expensive or destructive command merely because it appears in this list. Discover the project's documented commands first.

## 3. GitHub backlog to investigate

The public repository pages showed **11 open pull requests** and **23 open issues** when this plan was prepared. The listing may change; refresh it before implementation.

### Open pull requests

Review all open PRs, not just the newest one:

- [ ] PR #194 — [Add User/Group subjects to generated RBAC bindings](https://github.com/sighupio/permission-manager/pull/194) (draft; opened 2025-11-29). Inspect the patch, intended API/configuration, tests, generated RBAC, and compatibility. Decide whether to improve it, split it, or leave it untouched. Do not silently duplicate or overwrite contributor work.
- [ ] PR #183 — [Fix indents in CRD definition](https://github.com/sighupio/permission-manager/pull/183).
- [ ] PR #182 — [Update validUsernameRegex](https://github.com/sighupio/permission-manager/pull/182).
- [ ] PR #180, #178, #177, #176 — Dependabot `word-wrap` / `semver` updates in UI and e2e UI. Check whether these are still relevant, whether lockfiles are current, and whether the dependency is still present.
- [ ] PR #179 — [Replace Cypress with Playwright](https://github.com/sighupio/permission-manager/pull/179). Assess migration completeness, CI, test coverage, and maintenance cost.
- [ ] PR #172 — [Add secret dynamically](https://github.com/sighupio/permission-manager/pull/172). Review secret lifecycle, validation, RBAC, and security implications.
- [ ] PR #166 — [Update UI stack](https://github.com/sighupio/permission-manager/pull/166).
- [ ] PR #165 — [Evolve frontend deployment](https://github.com/sighupio/permission-manager/pull/165).

For every PR, record: status, changed areas, review comments, whether its branch can be fetched, overlap with other work, test status, and recommendation. Do not merge PRs.

### Open issues

Investigate these issues and reproduce them against the current code before deciding they need fixes:

| Priority | Issue | Investigation target |
|---|---|---|
| Fixed | [#193 Kubernetes v1.34](https://github.com/sighupio/permission-manager/issues/193) | Establish supported Kubernetes range; run compatibility tests against v1.34; find removed/deprecated APIs and dependency incompatibilities. |
| Fixed | [#168 Kubeconfig generation on Kubernetes >1.24](https://github.com/sighupio/permission-manager/issues/168) | Investigate service-account token behavior after Kubernetes 1.24; avoid assuming auto-created long-lived token Secrets; design a supported, secure token flow. |
| Fixed | [#186 `read` is not a valid verb](https://github.com/sighupio/permission-manager/issues/186) | Trace verb validation and Kubernetes RBAC verb semantics; determine whether this is a UI, API, validation, or documentation issue. |
| Medium | [#191 Create permission-manager through API or command](https://github.com/sighupio/permission-manager/issues/191) | Determine whether a supported API/CLI exists or is feasible; define authentication, authorization, validation, idempotency, and lifecycle semantics before implementation. |
| Medium | [#190 Make basic auth optional](https://github.com/sighupio/permission-manager/issues/190) | Trace authentication middleware/configuration; ensure disabling basic auth does not accidentally expose the UI/API and document ingress/OIDC assumptions. |
| Medium | [#171 Use an existing Secret in Helm chart](https://github.com/sighupio/permission-manager/issues/171) | Check chart values/templates and secret references; avoid rendering secret data into Helm release metadata. |
| Medium | [#170 Allow annotations in Helm chart Secret](https://github.com/sighupio/permission-manager/issues/170) | Verify templating, annotation merge behavior, nil handling, and rendered output tests. |
| Medium | [#169 Wrong Helm chart appVersion format](https://github.com/sighupio/permission-manager/issues/169) | Verify chart metadata and release/version automation; establish the expected SemVer format. |
| Medium | [#189 Index the Helm chart on ArtifactHub](https://github.com/sighupio/permission-manager/issues/189) | Verify current chart repository, Artifact Hub metadata, publishing workflow, and whether the issue remains relevant. |
| Fixed | [#173 Kubernetes version confirmation](https://github.com/sighupio/permission-manager/issues/173) | Reconcile reported version with chart metadata, docs, CI, and dependencies; update one canonical compatibility statement. |
| Low | [#157 Improve `rename-helm-release.sh`](https://github.com/sighupio/permission-manager/issues/157) | Review quoting, error handling, idempotency, shell portability, and tests. |

Also inspect the remaining open issues and all issue comments. The table is a starting shortlist, not a complete backlog.

## 4. Suggested execution order

Do not implement every item in one giant patch. Create a separate branch or commit-sized change for each approved task.

### Phase A — Establish support and compatibility

- [ ] Determine the currently supported Kubernetes versions.
- [ ] Build a CI/test matrix that covers the declared minimum and maximum supported versions, including v1.34 if feasible.
- [ ] Investigate issue #193 and #173 together to avoid conflicting version claims.
- [ ] Audit Kubernetes API usage, client libraries, controller-runtime (if used), generated clients, CRDs, and Helm templates.
- [ ] Update compatibility documentation only after the evidence is clear.

**Acceptance:** version support is explicit, reproducible in CI where feasible, and consistent across docs, chart metadata, and release artifacts.

### Phase B — Kubeconfig and service-account token correctness

- [ ] Reproduce issue #168 on a disposable Kubernetes cluster newer than v1.24.
- [ ] Trace how permission-manager creates identities, ServiceAccounts, token Secrets, and kubeconfig output.
- [ ] Prefer the Kubernetes TokenRequest API / short-lived token flow where appropriate and supported by the project's design.
- [ ] Do not create long-lived token Secrets by default merely to preserve old behavior.
- [ ] Document token expiry, renewal, audience, permissions, and operational impact.
- [ ] Add tests for missing token Secret, delayed controller population, expired token, API errors, and supported cluster versions.

**Acceptance:** kubeconfig generation works on supported Kubernetes versions without relying on removed automatic token-Secret behavior; security characteristics are documented and tested.

### Phase C — RBAC semantics and User/Group subjects

- [ ] Reproduce issue #186 and identify where `read` is accepted/rejected.
- [ ] Confirm Kubernetes RBAC semantics; do not invent a `read` verb if the intended meaning is `get`, `list`, and `watch`.
- [ ] Review PR #194 for generated `User`/`Group` subjects and the intended data model.
- [ ] Check that RoleBinding and ClusterRoleBinding subjects are rendered correctly for `User`, `Group`, and `ServiceAccount`.
- [ ] Validate namespace/cluster scope and prevent privilege escalation through malformed subjects or role references.
- [ ] Add unit, API validation, and rendered-manifest tests.

**Acceptance:** supported subjects and verbs match Kubernetes semantics, validation is consistent across UI/API/backend, and generated RBAC is correct.

### Phase D — Authentication configuration

- [ ] Reproduce issue #190.
- [ ] Map all authentication paths (UI, API, websocket if present, health endpoints, static assets).
- [ ] Define secure behavior when basic auth is disabled; fail closed where configuration is ambiguous.
- [ ] Ensure there is a clear alternative trust boundary (for example, an explicitly configured external auth proxy) rather than silently allowing anonymous access.
- [ ] Add tests for enabled, disabled, invalid, and missing configuration.
- [ ] Document deployment examples and security warnings.

**Acceptance:** auth behavior is explicit, tested, and cannot be disabled accidentally by an empty or malformed value.

### Phase E — Helm chart and secret management

- [ ] Investigate #171, #170, #169, and #189 together.
- [ ] Add support for referencing an existing Secret if consistent with current chart design.
- [ ] Add configurable annotations without dropping chart-managed annotations.
- [ ] Validate `appVersion` and chart version conventions against release automation.
- [ ] Render charts with default, custom, and existing-secret configurations.
- [ ] Add `helm lint` and template assertions to CI if not already present.
- [ ] Verify Artifact Hub metadata and publishing process before changing release configuration.

**Acceptance:** templates render deterministically, existing Secrets are referenced rather than copied, metadata is valid, and chart behavior is documented.

### Phase F — API/CLI creation workflow

- [ ] Analyze issue #191 and existing API/CLI patterns.
- [ ] Write a small design note before implementing anything.
- [ ] Define authentication/authorization, input validation, idempotency, error responses, auditability, and compatibility.
- [ ] Reuse existing controller/domain logic rather than creating a second implementation path.
- [ ] Add tests for unauthorized requests, invalid input, duplicate resources, and successful creation.
- [ ] Implement only if maintainers' intended product direction is clear.

**Acceptance:** a documented, authorized, testable creation path exists, or a clear explanation is recorded for why it should remain out of scope.

### Phase G — Frontend, e2e, and dependency maintenance

- [ ] Review PR #179 and #166 before changing the UI stack or test framework.
- [ ] Determine the currently maintained frontend framework and package manager from lockfiles and CI.
- [ ] Avoid mixing framework migration, test-runner migration, and unrelated UI fixes.
- [ ] Reassess Dependabot PRs #176–#180 against current lockfiles and dependency usage.
- [ ] Add or preserve coverage for login, permission creation/editing, RBAC subjects, kubeconfig generation, and error states.
- [ ] Keep browser tests deterministic and avoid real-cluster dependencies unless the e2e suite explicitly provisions one.

**Acceptance:** frontend build, unit tests, and browser tests run reproducibly in CI with maintained dependencies.

### Phase H — Scripts, documentation, and project health

- [ ] Review issue #157 shell script behavior and add tests for paths with spaces, missing tools, invalid arguments, and failed commands.
- [ ] Investigate issue #192 ("is this project dead?") as a project-maintenance/documentation concern, not a code change to make speculatively.
- [ ] Document release cadence, supported versions, installation, upgrade, backup/recovery, and contribution workflow where information is missing.
- [ ] Ensure examples match actual chart values and API schemas.

## 5. Security review checklist

Permission-manager manages Kubernetes permissions. Treat security regressions as release blockers.

- [ ] Check for privilege escalation in RoleBinding/ClusterRoleBinding creation and updates.
- [ ] Verify namespace boundaries and cluster-scoped operations.
- [ ] Ensure a user cannot grant permissions beyond the application's intended authorization model.
- [ ] Check authentication defaults, session handling, CSRF protections (where applicable), and proxy trust configuration.
- [ ] Never log passwords, bearer tokens, kubeconfigs, Secret data, or session cookies.
- [ ] Avoid storing long-lived Kubernetes credentials unless explicitly required and protected.
- [ ] Validate all user-controlled names, namespaces, subjects, verbs, API groups, and resource names.
- [ ] Check YAML/Helm template injection and unsafe shell invocation.
- [ ] Review dependencies and container base images using existing project tooling.
- [ ] Add negative tests for unauthorized and malformed requests.

## 6. Test and validation strategy

Use the repository's actual commands, but aim for these layers:

1. **Unit tests:** affected Go packages and frontend packages.
2. **Static checks:** `go vet`, project lint tools, formatting, frontend lint/typecheck.
3. **Build:** all supported binaries and frontend bundles.
4. **API/CRD validation:** generated code consistency and CRD schema checks.
5. **Helm:** `helm lint` and `helm template` for representative values.
6. **Integration/e2e:** disposable Kubernetes clusters for supported versions; never a production cluster.
7. **Security:** authorization-negative tests and secret/token handling tests.
8. **Diff review:** `git diff --check`, inspect generated changes, and ensure no unrelated modifications.

A test that cannot run locally must be reported as **not run**, with the reason. Do not claim success based only on compilation.

## 7. Local Ollama / OpenCode operating guidance

Use the existing OpenCode Ollama provider configuration. Do not add cloud models or send source code to external inference services.

- Prefer a coding-capable local model with sufficient context for the repository.
- Keep temperature low for code changes.
- Use a large context window where the local model supports it, but provide focused files rather than dumping the entire repository into every prompt.
- Use separate analysis and implementation passes when practical.
- Ask the model to cite file paths and symbols from the checked-out repository for every diagnosis.
- Require tests and command output as evidence.
- If the model cannot reliably reason across a large change, split the task into smaller patches.
- Never let model-generated claims replace compiler, test, linter, or cluster evidence.

Example OpenCode task prompt:

```text
Read plan.md completely. First inspect repository instructions and git status.
Do not edit files yet.

Choose the highest-priority unfinished task that can be reproduced from the current
checkout. Inspect the relevant GitHub issue/PR and current implementation. Report:
1. current behavior and exact code path,
2. reproduction steps,
3. proposed minimal fix,
4. regression tests,
5. compatibility/security risks,
6. exact validation commands.

Wait for approval before making changes.
```

After approval:

```text
Implement only the approved task from plan.md. Add regression tests first where
practical. Do not modify unrelated files, do not commit or push, and do not change
public behavior beyond the approved scope. Run the focused tests and relevant
repository checks. Review the full diff and report changed files, test results,
remaining risks, and anything not tested.
```

## 8. Definition of done

A task is complete only when:

- [ ] The original behavior was reproduced or its current status was verified.
- [ ] The change is minimal and consistent with project conventions.
- [ ] Regression tests cover the reported behavior and relevant negative cases.
- [ ] Relevant tests, lint, and builds pass (or limitations are explicitly documented).
- [ ] Documentation and examples are updated where needed.
- [ ] Security and backwards-compatibility implications are reviewed.
- [ ] The final diff contains no unrelated changes or secrets.
- [ ] A concise report identifies the issue/PR, root cause, fix, tests, and remaining risks.

## 9. Final reporting format

For each completed task, report:

```text
Task:
Related issue/PR:
Status: complete | blocked | needs maintainer decision

Root cause:
Files changed:
Behavioral change:
Tests added/updated:
Commands executed:
Results:
Compatibility/security notes:
Known limitations:
Recommended next task:
```

## 10. Source links

- [Open pull requests](https://github.com/sighupio/permission-manager/pulls)
- [Open issues](https://github.com/sighupio/permission-manager/issues)
- [PR #194 — User/Group subjects](https://github.com/sighupio/permission-manager/pull/194)
- [Issue #193 — Kubernetes v1.34](https://github.com/sighupio/permission-manager/issues/193)
- [Issue #192 — Project maintenance](https://github.com/sighupio/permission-manager/issues/192)
- [Issue #191 — API/command creation](https://github.com/sighupio/permission-manager/issues/191)
- [Issue #190 — Optional basic auth](https://github.com/sighupio/permission-manager/issues/190)
- [Issue #189 — Artifact Hub](https://github.com/sighupio/permission-manager/issues/189)
- [Issue #186 — `read` verb](https://github.com/sighupio/permission-manager/issues/186)
- [Issue #173 — Kubernetes version](https://github.com/sighupio/permission-manager/issues/173)
- [Issue #171 — Existing Helm Secret](https://github.com/sighupio/permission-manager/issues/171)
- [Issue #170 — Secret annotations](https://github.com/sighupio/permission-manager/issues/170)
- [Issue #169 — Helm appVersion](https://github.com/sighupio/permission-manager/issues/169)
- [Issue #168 — Kubeconfig on Kubernetes >1.24](https://github.com/sighupio/permission-manager/issues/168)
- [Issue #157 — Helm release rename script](https://github.com/sighupio/permission-manager/issues/157)
