package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestIsLegacyBinding(t *testing.T) {
	user := rbacv1.Subject{Kind: "User", Name: "jane", APIGroup: "rbac.authorization.k8s.io"}
	sa := rbacv1.Subject{Kind: "ServiceAccount", Name: "jane", Namespace: "permission-manager"}
	other := rbacv1.Subject{Kind: "User", Name: "bob", APIGroup: "rbac.authorization.k8s.io"}

	assert.True(t, isLegacyBinding([]rbacv1.Subject{user}, "jane"), "user-only binding is legacy")
	assert.False(t, isLegacyBinding([]rbacv1.Subject{sa}, "jane"), "service account binding is not legacy")
	assert.False(t, isLegacyBinding([]rbacv1.Subject{sa, user}, "jane"), "binding generated with ADDITIONAL_SUBJECT_KINDS=User is not legacy")
	assert.False(t, isLegacyBinding([]rbacv1.Subject{other}, "jane"), "other users are ignored")
	assert.False(t, isLegacyBinding(nil, "jane"))
}

func TestRoleBindingCreateAndLegacyCheck(t *testing.T) {
	m, _ := newTestManager()

	_, err := m.RoleBindingCreate("dev", "jane", RoleBindingRequirements{
		RoleKind:        "ClusterRole",
		RoleName:        "template-namespaced-resources___developer",
		RolebindingName: "jane___developer___dev",
		Subjects:        []rbacv1.Subject{{Kind: "User", Name: "jane", APIGroup: "rbac.authorization.k8s.io"}},
	})
	require.NoError(t, err)

	rbs, err := m.RoleBindingList("dev")
	require.NoError(t, err)
	require.Len(t, rbs.Items, 1)
	assert.Equal(t, "jane", rbs.Items[0].Labels["generated_for_user"])
	assert.Equal(t, "rbac.authorization.k8s.io", rbs.Items[0].RoleRef.APIGroup)

	legacy, err := m.RoleBindingLegacyCheck("dev", "jane")
	require.NoError(t, err)
	require.NotNil(t, legacy)
	assert.Equal(t, "jane___developer___dev", legacy.Name)

	legacy, err = m.RoleBindingLegacyCheck("dev", "someone-else")
	require.NoError(t, err)
	assert.Nil(t, legacy)
}

func TestClusterRoleBindingCreateAndLegacyCheck(t *testing.T) {
	m, _ := newTestManager()

	subjects := []rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane", Namespace: "permission-manager"},
		{Kind: "User", Name: "jane", APIGroup: "rbac.authorization.k8s.io"},
	}
	crb, err := m.ClusterRoleBindingCreate("jane___admin", "jane", "template-cluster-resources___admin", subjects)
	require.NoError(t, err)
	assert.Equal(t, "ClusterRole", crb.RoleRef.Kind)
	assert.Equal(t, subjects, crb.Subjects)

	legacy, err := m.ClusterRoleBindingLegacyCheck("jane")
	require.NoError(t, err)
	assert.Nil(t, legacy, "bindings with the service account subject are not legacy")

	require.NoError(t, m.ClusterRoleBindingDelete("jane___admin"))
	list, err := m.ClusterRoleBindingList()
	require.NoError(t, err)
	assert.Empty(t, list.Items)
}
