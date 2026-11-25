package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sighupio/permission-manager/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func testConfig() config.Config {
	return config.Config{
		Cluster: config.ClusterConfig{
			Name:                "test-cluster",
			ControlPlaneAddress: "https://k8s.example.com:6443",
			Namespace:           "permission-manager",
		},
		Backend: config.BackendConfig{Port: "4000"},
		Auth: config.AuthConfig{
			BasicAuthEnabled: true,
			Username:         "admin",
			Password:         "secret",
		},
		Kubeconfig: config.KubeconfigConfig{TokenMode: config.TokenModeSecret},
	}
}

func doRequest(t *testing.T, cfg config.Config, client *fake.Clientset, method, path, body string, withAuth bool) *httptest.ResponseRecorder {
	t.Helper()
	e := NewWithClient(cfg, client)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if withAuth {
		req.SetBasicAuth(cfg.Auth.Username, cfg.Auth.Password)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestBasicAuthIsEnforced(t *testing.T) {
	rec := doRequest(t, testConfig(), fake.NewSimpleClientset(), http.MethodGet, "/api/list-namespace", "", false)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	cfg := testConfig()
	e := NewWithClient(cfg, fake.NewSimpleClientset())
	req := httptest.NewRequest(http.MethodGet, "/api/list-namespace", nil)
	req.SetBasicAuth("admin", "wrong")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = doRequest(t, cfg, fake.NewSimpleClientset(), http.MethodGet, "/api/list-namespace", "", true)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBasicAuthCanBeDisabled(t *testing.T) {
	cfg := testConfig()
	cfg.Auth = config.AuthConfig{BasicAuthEnabled: false}

	rec := doRequest(t, cfg, fake.NewSimpleClientset(), http.MethodGet, "/api/list-namespace", "", false)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateRoleBindingAddsConfiguredSubjects(t *testing.T) {
	cfg := testConfig()
	cfg.RBAC.AdditionalSubjectKinds = []string{"User"}
	client := fake.NewSimpleClientset()

	body := `{"generated_for_user":"jane","roleName":"template-namespaced-resources___developer","namespace":"dev","roleKind":"ClusterRole",` +
		`"subjects":[{"kind":"ServiceAccount","name":"jane","namespace":"kube-system"}],"rolebindingName":"jane___developer___dev"}`

	rec := doRequest(t, cfg, client, http.MethodPost, "/api/create-rolebinding", body, true)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rb, err := client.RbacV1().RoleBindings("dev").Get(context.Background(), "jane___developer___dev", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, []rbacv1.Subject{
		{Kind: "ServiceAccount", Name: "jane", Namespace: "permission-manager"},
		{Kind: "User", Name: "jane", APIGroup: "rbac.authorization.k8s.io"},
	}, rb.Subjects)
	assert.Equal(t, "jane", rb.Labels["generated_for_user"])
}

func TestCreateRoleBindingRejectsInvalidInput(t *testing.T) {
	tests := map[string]string{
		"bad role kind": `{"generated_for_user":"jane","roleName":"x","namespace":"dev","roleKind":"Pod",` +
			`"subjects":[{"kind":"ServiceAccount","name":"jane"}],"rolebindingName":"rb"}`,
		"bad subject kind": `{"generated_for_user":"jane","roleName":"x","namespace":"dev","roleKind":"ClusterRole",` +
			`"subjects":[{"kind":"Node","name":"jane"}],"rolebindingName":"rb"}`,
		"missing fields": `{"roleName":"x"}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			rec := doRequest(t, testConfig(), client, http.MethodPost, "/api/create-rolebinding", body, true)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			list, err := client.RbacV1().RoleBindings("").List(context.Background(), metav1.ListOptions{})
			require.NoError(t, err)
			assert.Empty(t, list.Items)
		})
	}
}

func TestCreateClusterRoleBindingUsesGeneratedForUserLabel(t *testing.T) {
	client := fake.NewSimpleClientset()

	body := `{"generated_for_user":"jane","roleName":"template-cluster-resources___read-only",` +
		`"subjects":[{"kind":"ServiceAccount","name":"jane","namespace":"permission-manager"}],` +
		`"clusterRolebindingName":"jane___template-cluster-resources___read-only"}`

	rec := doRequest(t, testConfig(), client, http.MethodPost, "/api/create-cluster-rolebinding", body, true)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	crb, err := client.RbacV1().ClusterRoleBindings().Get(context.Background(), "jane___template-cluster-resources___read-only", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "jane", crb.Labels["generated_for_user"])
}

func TestCreateKubeconfigRejectsInvalidUsername(t *testing.T) {
	rec := doRequest(t, testConfig(), fake.NewSimpleClientset(), http.MethodPost, "/api/create-kubeconfig", `{"username":"../etc"}`, true)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateKubeconfigReturnsErrorInsteadOfEmptyToken(t *testing.T) {
	// nothing populates the token Secret with the fake client, so generation must fail
	// with a server error instead of returning a kubeconfig with an empty token (issue #168)
	if testing.Short() {
		t.Skip("waits for the token timeout")
	}

	rec := doRequest(t, testConfig(), fake.NewSimpleClientset(), http.MethodPost, "/api/create-kubeconfig", `{"username":"jane","namespace":"dev"}`, true)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Contains(t, res["error"], "kubeconfig generation failed")
}
