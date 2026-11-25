package resources

import (
	"context"
	b64 "encoding/base64"
	"testing"
	"time"

	"sighupio/permission-manager/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

const testNamespace = "permission-manager"

var testCluster = config.ClusterConfig{
	Name:                "My-cluster",
	ControlPlaneAddress: "https://100.200.10.200",
	Namespace:           testNamespace,
}

func newTestManager(objects ...runtime.Object) (*Manager, *fake.Clientset) {
	client := fake.NewSimpleClientset(objects...)
	m := NewManager(client, context.Background())
	m.tokenWaitTimeout = 200 * time.Millisecond
	m.tokenPollInterval = 10 * time.Millisecond
	return m, client
}

func expectedKubeconfig(ca, token string) string {
	return `---
apiVersion: v1
kind: Config
current-context: john.doe@My-cluster
clusters:
  - cluster:
      certificate-authority-data: ` + b64.StdEncoding.EncodeToString([]byte(ca)) + `
      server: https://100.200.10.200
    name: My-cluster
contexts:
  - context:
      cluster: My-cluster
      user: john.doe
      namespace: dev
    name: john.doe@My-cluster
users:
  - name: john.doe
    user:
      token: ` + token
}

func tokenSecret(name, uid string, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
			Annotations: map[string]string{
				v1.ServiceAccountNameKey: name,
				v1.ServiceAccountUIDKey:  uid,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
		Data: data,
	}
}

func TestCreateKubeconfigFromExistingSecret(t *testing.T) {
	sa := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "john.doe", Namespace: testNamespace, UID: "uid-1"}}
	secret := tokenSecret("john.doe", "uid-1", map[string][]byte{"token": []byte("TOKEN"), "ca.crt": []byte("CA")})

	m, _ := newTestManager(sa, secret)

	got, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeSecret}, "john.doe", "dev")

	require.NoError(t, err)
	assert.Equal(t, expectedKubeconfig("CA", "TOKEN"), got)
}

func TestCreateKubeconfigCreatesServiceAccountAndSecret(t *testing.T) {
	m, client := newTestManager()

	// simulate the token controller populating the Secret shortly after it is created
	go func() {
		time.Sleep(30 * time.Millisecond)
		for i := 0; i < 50; i++ {
			s, err := client.CoreV1().Secrets(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
			if err == nil {
				s.Data = map[string][]byte{"token": []byte("TOKEN"), "ca.crt": []byte("CA")}
				_, _ = client.CoreV1().Secrets(testNamespace).Update(context.Background(), s, metav1.UpdateOptions{})
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	got, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeSecret}, "john.doe", "dev")
	require.NoError(t, err)
	assert.Equal(t, expectedKubeconfig("CA", "TOKEN"), got)

	_, err = client.CoreV1().ServiceAccounts(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
	assert.NoError(t, err, "service account must be created")

	s, err := client.CoreV1().Secrets(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, v1.SecretTypeServiceAccountToken, s.Type)
	assert.Equal(t, "john.doe", s.Annotations[v1.ServiceAccountNameKey])
}

func TestCreateKubeconfigFallsBackToRootCAConfigMap(t *testing.T) {
	sa := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "john.doe", Namespace: testNamespace, UID: "uid-1"}}
	secret := tokenSecret("john.doe", "uid-1", map[string][]byte{"token": []byte("TOKEN")})
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "kube-root-ca.crt", Namespace: testNamespace},
		Data:       map[string]string{"ca.crt": "ROOTCA"},
	}

	m, _ := newTestManager(sa, secret, cm)

	got, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeSecret}, "john.doe", "dev")
	require.NoError(t, err)
	assert.Equal(t, expectedKubeconfig("ROOTCA", "TOKEN"), got)
}

func TestCreateKubeconfigTimesOutWhenTokenNeverPopulated(t *testing.T) {
	m, _ := newTestManager()

	got, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeSecret}, "john.doe", "dev")

	// before issue #168 was fixed an (invalid) kubeconfig with an empty token was returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
	assert.Empty(t, got)
}

func TestCreateKubeconfigRecreatesStaleSecret(t *testing.T) {
	sa := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "john.doe", Namespace: testNamespace, UID: "new-uid"}}
	stale := tokenSecret("john.doe", "old-uid", map[string][]byte{"token": []byte("OLD"), "ca.crt": []byte("CA")})

	m, client := newTestManager(sa, stale)

	_, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeSecret}, "john.doe", "dev")
	// nothing populates the new secret in this test
	assert.Error(t, err)

	s, err := client.CoreV1().Secrets(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "new-uid", s.Annotations[v1.ServiceAccountUIDKey])
	assert.Empty(t, s.Data["token"], "the stale token must not be reused")
}

func TestCreateKubeconfigRejectsNonTokenSecret(t *testing.T) {
	sa := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "john.doe", Namespace: testNamespace}}
	opaque := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "john.doe", Namespace: testNamespace},
		Type:       v1.SecretTypeOpaque,
		Data:       map[string][]byte{"token": []byte("NOT-A-SA-TOKEN")},
	}

	m, _ := newTestManager(sa, opaque)

	_, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeSecret}, "john.doe", "dev")
	assert.Error(t, err)
}

func TestCreateKubeconfigWithTokenRequest(t *testing.T) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "kube-root-ca.crt", Namespace: testNamespace},
		Data:       map[string]string{"ca.crt": "ROOTCA"},
	}
	m, client := newTestManager(cm)

	var requestedExpiration int64
	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "token" {
			return false, nil, nil
		}
		tr := action.(k8stesting.CreateAction).GetObject().(*authenticationv1.TokenRequest)
		requestedExpiration = *tr.Spec.ExpirationSeconds
		return true, &authenticationv1.TokenRequest{Status: authenticationv1.TokenRequestStatus{Token: "BOUND-TOKEN"}}, nil
	})

	got, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{
		TokenMode:       config.TokenModeTokenRequest,
		TokenExpiration: 24 * time.Hour,
	}, "john.doe", "dev")

	require.NoError(t, err)
	assert.Equal(t, expectedKubeconfig("ROOTCA", "BOUND-TOKEN"), got)
	assert.Equal(t, int64(86400), requestedExpiration)

	_, err = client.CoreV1().Secrets(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err), "token request mode must not create long-lived secrets")
}

func TestCreateKubeconfigWithTokenRequestWithoutCA(t *testing.T) {
	m, client := newTestManager()
	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "token" {
			return false, nil, nil
		}
		return true, &authenticationv1.TokenRequest{Status: authenticationv1.TokenRequestStatus{Token: "BOUND-TOKEN"}}, nil
	})

	_, err := m.ServiceAccountCreateKubeConfigForUser(testCluster, config.KubeconfigConfig{TokenMode: config.TokenModeTokenRequest, TokenExpiration: time.Hour}, "john.doe", "dev")
	assert.Error(t, err)
}

func TestServiceAccountDelete(t *testing.T) {
	sa := &v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "john.doe", Namespace: testNamespace}}
	secret := tokenSecret("john.doe", "", nil)
	m, client := newTestManager(sa, secret)

	require.NoError(t, m.ServiceAccountDelete(testNamespace, "john.doe"))

	_, err := client.CoreV1().ServiceAccounts(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))
	_, err = client.CoreV1().Secrets(testNamespace).Get(context.Background(), "john.doe", metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err))

	// deleting again is a no-op
	assert.NoError(t, m.ServiceAccountDelete(testNamespace, "john.doe"))
}
