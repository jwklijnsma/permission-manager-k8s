package resources

import (
	"context"
	"time"

	k8sclient "k8s.io/client-go/kubernetes"
)

const (
	defaultTokenWaitTimeout  = 30 * time.Second
	defaultTokenPollInterval = 500 * time.Millisecond
)

// Manager allows to list and manage the life-cycle of the various K8s cluster resources managed by the PermissionManager.
type Manager struct {
	kubeclient                    k8sclient.Interface
	context                       context.Context
	V1Alpha1PermissionManagerUser V1Alpha1PermissionManagerUser

	// tokenWaitTimeout is how long to wait for the token controller to populate a token Secret
	tokenWaitTimeout time.Duration
	// tokenPollInterval is the polling interval used while waiting for a token Secret
	tokenPollInterval time.Duration
}

// NewManager returns a new instance of a ResourceService
// allowing to interact with a K8s cluster via the given K8s client interface.
func NewManager(kc k8sclient.Interface, ctx context.Context) *Manager {
	return &Manager{
		kubeclient:                    kc,
		context:                       ctx,
		V1Alpha1PermissionManagerUser: V1Alpha1PermissionManagerUser{kubeclient: kc, context: ctx},
		tokenWaitTimeout:              defaultTokenWaitTimeout,
		tokenPollInterval:             defaultTokenPollInterval,
	}
}
