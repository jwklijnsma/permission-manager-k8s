# Permission manager

![flow](./docs/assets/flow.gif)

Welcome to the **Permission Manager**! :tada: :tada:

[![ci](https://github.com/jwklijnsma/permission-manager-k8s/actions/workflows/ci.yml/badge.svg)](https://github.com/jwklijnsma/permission-manager-k8s/actions/workflows/ci.yml)
[![docs](https://github.com/jwklijnsma/permission-manager-k8s/actions/workflows/docs.yml/badge.svg)](https://jwklijnsma.github.io/permission-manager-k8s/)
![License](https://img.shields.io/github/license/jwklijnsma/permission-manager-k8s)

> Maintained fork of [sighupio/permission-manager](https://github.com/sighupio/permission-manager).
> Documentation: <https://jwklijnsma.github.io/permission-manager-k8s/> -
> status of the upstream issues and PRs: [docs/reference/upstream-backlog.md](docs/reference/upstream-backlog.md).

Permission Manager is an application developed by [SIGHUP](https://sighup.io) that enables a super-easy and user-friendly **RBAC management for Kubernetes**. If you are looking for a simple and intuitive way of managing your users within a Kubernetes cluster, this is the right place.

With Permission Manager, you can create users, assign namespaces/permissions, and distribute Kubeconfig YAML files via a nice&easy web UI.

## Screenshots

### First Page

![First Page](docs/assets/first-page.png)

### Creating a user

![Creating a user](docs/assets/create-user.png)

### Creating a user - Summary

![Create user Sumary](docs/assets/create-user-summary.png)

### User's Kubeconfig

![User's Kubeconfig](docs/assets/users-kubeconfig.png)

### Deleting a user

![Deleting a user](docs/assets/delete-user.png)

## Installation

To deploy and run the Permission Manager on your cluster, follow the [installation guide](docs/getting-started/install.md).

```shell
helm upgrade --install permission-manager oci://ghcr.io/jwklijnsma/charts/permission-manager \
  --namespace permission-manager --create-namespace \
  --set config.clusterName=my-cluster \
  --set config.controlPlaneAddress=https://my-cluster.example.com:6443 \
  --set config.basicAuthPassword=changeMe
```

## Development Setup

We love contributions from the community :heart:.
Learn [how-to-contribute](docs/development/contributing.md)

## FAQ

### How it works

The application allows us to define and select permission standard templates (those defining who can do what, within a namespace or globally) and associate them with all the users you might want to create.

The template system is an abstraction over Cluster-Roles, RoleBinding, and ClusterRolesBindigs.

### What is a template

A template is a ClusterRole with a prefix

`template-namespaced-resources___`

for example
`template-namespaced-resources___developer`

#### Why a template is not a CRD

When we started the development of this project, a template was one-to-one to a `clusterrole`, the usage of a CRD looked overkill. This will obviously change in future to avoid polluting `clusterroles` and will allow us to have a more precise encapsulation of what is owned by the permission manager.

### How to add a new template

Create a clusterrole starting with `template-namespaced-resources___` and apply it.

#### Default templates

`developer` and `operation` default templates are located at deployments/kubernetes/seeds/seed.yml (and `config.templates` in the Helm chart)

### What is a user

A user is a custom resource of kind `permissionmanagerusers.permissionmanager.user`

## Commercial & Enterprise Support

Are you using Permission Manager in your clusters and looking for commercial support? Let us know and [get in touch](mailto:sales@sighup.io)
