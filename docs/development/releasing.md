# Releasing

## Application image

```shell
bumpversion minor     # updates deploy.yml, Chart.yaml (appVersion) and the install docs, commits and tags vX.Y.Z
git push && git push --tags
```

The `docker` workflow builds the tag and pushes
`ghcr.io/jwklijnsma/permission-manager-k8s:vX.Y.Z` (plus `X.Y.Z` and `X.Y`).
The chart's `image.tag` defaults to `v` + `appVersion`, so bump `appVersion`
and push the `v*` tag before releasing a chart that points to it.

## Helm chart

1. Bump `version` in `helm_chart/Chart.yaml` (and `appVersion` if the image changed),
   update `artifacthub.io/changes`.
2. Tag and push:

    ```shell
    git tag helm-chart-v0.2.0 && git push origin helm-chart-v0.2.0
    ```

3. The `helm-release` workflow checks that the tag matches `Chart.yaml`, pushes
   `oci://ghcr.io/jwklijnsma/charts/permission-manager:0.2.0`, pushes
   `helm_chart/artifacthub-repo.yml` as `permission-manager:artifacthub.io`, and
   creates the GitHub release "Helm Chart v0.2.0".

## Artifact Hub

One time (upstream issue #189):

1. Make the `charts/permission-manager` package on GHCR public.
2. On <https://artifacthub.io> add a repository of kind *Helm charts* with the URL
   `oci://ghcr.io/jwklijnsma/charts/permission-manager`.
3. Copy the repository ID into `helm_chart/artifacthub-repo.yml` and release the
   chart again for the verified publisher badge.

## Docs

The site is published from `main` by the `docs` workflow. Enable GitHub Pages
once: *Settings → Pages → Source: GitHub Actions*.
