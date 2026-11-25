package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestNormalizeSubjectsPinsServiceAccountNamespace(t *testing.T) {
	got, err := normalizeSubjects([]rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane", Namespace: "kube-system"},
	}, "permission-manager", nil)

	assert.NoError(t, err)
	assert.Equal(t, []rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane", Namespace: "permission-manager"},
	}, got)
}

func TestNormalizeSubjectsAddsAdditionalKinds(t *testing.T) {
	got, err := normalizeSubjects([]rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane", Namespace: "permission-manager"},
	}, "permission-manager", []string{"User", "Group"})

	assert.NoError(t, err)
	assert.Equal(t, []rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane", Namespace: "permission-manager"},
		{Kind: "User", Name: "jane", APIGroup: "rbac.authorization.k8s.io"},
		{Kind: "Group", Name: "jane", APIGroup: "rbac.authorization.k8s.io"},
	}, got)
}

func TestNormalizeSubjectsUserAndGroup(t *testing.T) {
	got, err := normalizeSubjects([]rbacv1.Subject{
		{Kind: "User", Name: "jane@example.com", Namespace: "should-be-dropped"},
		{Kind: "Group", Name: "devs"},
	}, "permission-manager", nil)

	assert.NoError(t, err)
	assert.Equal(t, []rbacv1.Subject{
		{Kind: "User", Name: "jane@example.com", APIGroup: "rbac.authorization.k8s.io"},
		{Kind: "Group", Name: "devs", APIGroup: "rbac.authorization.k8s.io"},
	}, got)
}

func TestNormalizeSubjectsDeduplicates(t *testing.T) {
	// the client already sends a User subject (e.g. web-client from PR #194) and the
	// server is configured to add one as well: it must appear only once
	got, err := normalizeSubjects([]rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane"},
		{Kind: "User", Name: "jane", APIGroup: "rbac.authorization.k8s.io"},
		{Kind: "ServiceAccount", Name: "jane"},
	}, "permission-manager", []string{"User"})

	assert.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestNormalizeSubjectsRejectsInvalid(t *testing.T) {
	_, err := normalizeSubjects([]rbacv1.Subject{{Kind: "Robot", Name: "x"}}, "pm", nil)
	assert.Error(t, err)

	_, err = normalizeSubjects([]rbacv1.Subject{{Kind: "ServiceAccount", Name: ""}}, "pm", nil)
	assert.Error(t, err)

	_, err = normalizeSubjects(nil, "pm", nil)
	assert.Error(t, err)
}

func TestIsValidRoleKind(t *testing.T) {
	assert.True(t, isValidRoleKind("ClusterRole"))
	assert.True(t, isValidRoleKind("Role"))
	assert.False(t, isValidRoleKind("clusterrole"))
	assert.False(t, isValidRoleKind(""))
}
