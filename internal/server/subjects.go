package server

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
)

// normalizeSubjects validates and normalizes the subjects of a RoleBinding / ClusterRoleBinding
// requested through the API.
//
//   - ServiceAccount subjects are always pinned to the permission-manager namespace
//     (workaround for https://github.com/sighupio/permission-manager/issues/140), so a request
//     cannot bind ServiceAccounts living in other namespaces.
//   - User and Group subjects get the RBAC apiGroup and no namespace, as required by Kubernetes.
//   - Any other kind is rejected.
//   - For every ServiceAccount subject, a subject of each kind listed in additionalKinds
//     ("User", "Group") with the same name is added. This allows the generated bindings to also
//     apply to users authenticated through OIDC / client certificates (issues #117, #41, #91, PR #194).
//   - Duplicates are removed, order is preserved.
func normalizeSubjects(subjects []rbacv1.Subject, serviceAccountNamespace string, additionalKinds []string) ([]rbacv1.Subject, error) {
	out := make([]rbacv1.Subject, 0, len(subjects)*(1+len(additionalKinds)))
	seen := map[string]bool{}

	add := func(s rbacv1.Subject) {
		key := s.Kind + "/" + s.Namespace + "/" + s.Name
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, s)
	}

	var serviceAccountNames []string

	for _, s := range subjects {
		if s.Name == "" {
			return nil, fmt.Errorf("subject name cannot be empty")
		}

		switch s.Kind {
		case rbacv1.ServiceAccountKind:
			add(rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: s.Name, Namespace: serviceAccountNamespace})
			serviceAccountNames = append(serviceAccountNames, s.Name)
		case rbacv1.UserKind, rbacv1.GroupKind:
			add(rbacv1.Subject{Kind: s.Kind, Name: s.Name, APIGroup: rbacv1.GroupName})
		default:
			return nil, fmt.Errorf("unsupported subject kind %q: allowed kinds are ServiceAccount, User and Group", s.Kind)
		}
	}

	for _, name := range serviceAccountNames {
		for _, kind := range additionalKinds {
			add(rbacv1.Subject{Kind: kind, Name: name, APIGroup: rbacv1.GroupName})
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("at least one subject is required")
	}

	return out, nil
}

// isValidRoleKind reports whether kind can be referenced by a RoleBinding.
func isValidRoleKind(kind string) bool {
	return kind == "ClusterRole" || kind == "Role"
}
