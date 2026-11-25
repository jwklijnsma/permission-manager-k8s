package resources

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"time"

	"sighupio/permission-manager/internal/config"

	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// rootCAConfigMapName is the ConfigMap published by kube-controller-manager in every namespace
// (Kubernetes >= 1.20) containing the CA bundle used to verify the API server.
const rootCAConfigMapName = "kube-root-ca.crt"

func (r *Manager) ServiceAccountGet(namespace, name string) (*v1.ServiceAccount, error) {
	return r.kubeclient.CoreV1().ServiceAccounts(namespace).Get(r.context, name, metav1.GetOptions{})
}

func (r *Manager) ServiceAccountCreate(namespace, name string) (*v1.ServiceAccount, error) {
	return r.kubeclient.CoreV1().ServiceAccounts(namespace).Create(r.context, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
}

// ServiceAccountDelete removes the ServiceAccount of a user and its token Secret (if any).
// Deleting the ServiceAccount invalidates every token issued for it, so kubeconfigs that were
// handed out for a deleted user stop working. Missing objects are not an error.
func (r *Manager) ServiceAccountDelete(namespace, name string) error {
	err := r.kubeclient.CoreV1().Secrets(namespace).Delete(r.context, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting token secret %s/%s: %w", namespace, name, err)
	}

	err = r.kubeclient.CoreV1().ServiceAccounts(namespace).Delete(r.context, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting service account %s/%s: %w", namespace, name, err)
	}

	return nil
}

// ServiceAccountCreateKubeConfigForUser ensures a ServiceAccount exists for the user and returns a
// kubeconfig authenticating as that ServiceAccount.
//
// Since Kubernetes 1.24 token Secrets are no longer generated automatically for ServiceAccounts
// (issue #168). Depending on kc.TokenMode the token is obtained either from an explicitly created
// "kubernetes.io/service-account-token" Secret (non-expiring, default) or from the TokenRequest
// API (time-bound, kc.TokenExpiration).
func (r *Manager) ServiceAccountCreateKubeConfigForUser(cluster config.ClusterConfig, kc config.KubeconfigConfig, username, kubeConfigNamespace string) (string, error) {
	serviceAccountNamespace := cluster.Namespace

	serviceAccount, err := r.ensureServiceAccount(serviceAccountNamespace, username)
	if err != nil {
		return "", err
	}

	var token string
	var ca []byte

	switch kc.TokenMode {
	case config.TokenModeTokenRequest:
		token, err = r.requestToken(serviceAccountNamespace, username, kc.TokenExpiration)
		if err != nil {
			return "", err
		}
	default:
		token, ca, err = r.serviceAccountSecretToken(serviceAccountNamespace, serviceAccount)
		if err != nil {
			return "", err
		}
	}

	if len(ca) == 0 {
		ca, err = r.clusterCA(serviceAccountNamespace)
		if err != nil {
			return "", err
		}
	}

	return renderKubeconfig(cluster, username, kubeConfigNamespace, ca, token), nil
}

func (r *Manager) ensureServiceAccount(namespace, name string) (*v1.ServiceAccount, error) {
	sa, err := r.ServiceAccountGet(namespace, name)
	if err == nil {
		return sa, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting service account %s/%s: %w", namespace, name, err)
	}

	sa, err = r.ServiceAccountCreate(namespace, name)
	if apierrors.IsAlreadyExists(err) {
		return r.ServiceAccountGet(namespace, name)
	}
	if err != nil {
		return nil, fmt.Errorf("creating service account %s/%s: %w", namespace, name, err)
	}

	return sa, nil
}

// serviceAccountSecretToken returns the token (and CA) stored in the token Secret of the
// ServiceAccount, creating the Secret when missing and waiting for the token controller to
// populate it.
func (r *Manager) serviceAccountSecretToken(namespace string, sa *v1.ServiceAccount) (string, []byte, error) {
	name := sa.GetName()

	secret, err := r.getSecretIfExists(namespace, name)
	if err != nil {
		return "", nil, err
	}

	// the Secret belongs to a previous incarnation of the ServiceAccount: recreate it
	if secret != nil && sa.GetUID() != "" && secret.Annotations[v1.ServiceAccountUIDKey] != string(sa.GetUID()) {
		log.Printf("token secret %s/%s is bound to another service account uid, recreating it", namespace, name)
		err = r.kubeclient.CoreV1().Secrets(namespace).Delete(r.context, name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return "", nil, fmt.Errorf("deleting stale token secret %s/%s: %w", namespace, name, err)
		}
		secret = nil
	}

	if secret != nil && secret.Type != v1.SecretTypeServiceAccountToken {
		return "", nil, fmt.Errorf("secret %s/%s exists but is not of type %s", namespace, name, v1.SecretTypeServiceAccountToken)
	}

	if secret == nil {
		_, err = r.SecretCreate(namespace, &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				// the token controller fills the Secret only if name and uid match the ServiceAccount
				Annotations: map[string]string{
					v1.ServiceAccountNameKey: name,
					v1.ServiceAccountUIDKey:  string(sa.GetUID()),
				},
			},
			Type: v1.SecretTypeServiceAccountToken,
		})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return "", nil, fmt.Errorf("creating token secret %s/%s: %w", namespace, name, err)
		}
	}

	deadline := time.Now().Add(r.tokenWaitTimeout)
	for {
		secret, err = r.getSecretIfExists(namespace, name)
		if err != nil {
			return "", nil, err
		}

		if secret != nil && len(secret.Data[v1.ServiceAccountTokenKey]) > 0 {
			return string(secret.Data[v1.ServiceAccountTokenKey]), secret.Data[v1.ServiceAccountRootCAKey], nil
		}

		if err := r.context.Err(); err != nil {
			return "", nil, err
		}

		if time.Now().After(deadline) {
			return "", nil, fmt.Errorf("timed out after %s waiting for the token of secret %s/%s to be populated "+
				"(is the service account token controller running?)", r.tokenWaitTimeout, namespace, name)
		}

		time.Sleep(r.tokenPollInterval)
	}
}

func (r *Manager) getSecretIfExists(namespace, name string) (*v1.Secret, error) {
	secret, err := r.SecretGet(namespace, name)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting secret %s/%s: %w", namespace, name, err)
	}
	return secret, nil
}

// requestToken mints a time-bound token for the ServiceAccount using the TokenRequest API.
func (r *Manager) requestToken(namespace, name string, expiration time.Duration) (string, error) {
	expirationSeconds := int64(expiration.Seconds())

	tr, err := r.kubeclient.CoreV1().ServiceAccounts(namespace).CreateToken(r.context, name, &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &expirationSeconds,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("requesting token for service account %s/%s: %w", namespace, name, err)
	}

	if tr == nil || tr.Status.Token == "" {
		return "", fmt.Errorf("the TokenRequest API returned an empty token for service account %s/%s", namespace, name)
	}

	return tr.Status.Token, nil
}

// clusterCA returns the API server CA bundle published in the kube-root-ca.crt ConfigMap.
func (r *Manager) clusterCA(namespace string) ([]byte, error) {
	cm, err := r.kubeclient.CoreV1().ConfigMaps(namespace).Get(r.context, rootCAConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting configmap %s/%s: %w", namespace, rootCAConfigMapName, err)
	}

	ca := cm.Data[v1.ServiceAccountRootCAKey]
	if ca == "" {
		return nil, fmt.Errorf("configmap %s/%s has no %s key", namespace, rootCAConfigMapName, v1.ServiceAccountRootCAKey)
	}

	return []byte(ca), nil
}

func renderKubeconfig(cluster config.ClusterConfig, username, namespace string, ca []byte, token string) string {
	tpl := `---
apiVersion: v1
kind: Config
current-context: %s@%s
clusters:
  - cluster:
      certificate-authority-data: %s
      server: %s
    name: %s
contexts:
  - context:
      cluster: %s
      user: %s
      namespace: %s
    name: %s@%s
users:
  - name: %s
    user:
      token: %s`

	return fmt.Sprintf(tpl,
		username,
		cluster.Name,
		b64.StdEncoding.EncodeToString(ca),
		cluster.ControlPlaneAddress,
		cluster.Name,
		cluster.Name,
		username,
		namespace,
		username,
		cluster.Name,
		username,
		token,
	)
}
