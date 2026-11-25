package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Token modes used when generating a user kubeconfig.
const (
	// TokenModeSecret creates (or reuses) a "kubernetes.io/service-account-token" Secret
	// bound to the user ServiceAccount. The token does not expire. This is the historical behaviour.
	TokenModeSecret = "secret"
	// TokenModeTokenRequest uses the TokenRequest API to mint a time-bound token.
	TokenModeTokenRequest = "tokenrequest"

	// DefaultTokenExpiration is used by TokenModeTokenRequest when no expiration is configured.
	DefaultTokenExpiration = 8760 * time.Hour
	// MinTokenExpiration is the minimum expiration accepted by the Kubernetes TokenRequest API.
	MinTokenExpiration = 10 * time.Minute

	DefaultBasicAuthUsername = "admin"
)

type ClusterConfig struct {
	Name                string
	ControlPlaneAddress string
	Namespace           string
}

type BackendConfig struct {
	Port string
}

// AuthConfig holds the UI/API authentication settings.
type AuthConfig struct {
	// BasicAuthEnabled is true unless BASIC_AUTH_DISABLED is explicitly set to a true value.
	BasicAuthEnabled bool
	Username         string
	Password         string
}

// RBACConfig holds the settings used when generating RoleBindings / ClusterRoleBindings.
type RBACConfig struct {
	// AdditionalSubjectKinds lists extra subject kinds ("User", "Group") that are added, with the
	// permission-manager username as name, to every binding generated for a user.
	AdditionalSubjectKinds []string
}

// KubeconfigConfig holds the settings used when generating user kubeconfigs.
type KubeconfigConfig struct {
	TokenMode       string
	TokenExpiration time.Duration
}

// Config contains PermissionManager cluster/server configuration
type Config struct {
	Cluster    ClusterConfig
	Backend    BackendConfig
	Auth       AuthConfig
	RBAC       RBACConfig
	Kubeconfig KubeconfigConfig
}

// New loads the configuration from the environment and terminates the process if it is invalid.
func New() *Config {
	cfg, err := Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}

	if !cfg.Auth.BasicAuthEnabled {
		log.Print("WARNING: basic auth is disabled (BASIC_AUTH_DISABLED=true). " +
			"Permission Manager must be protected by an authenticating proxy, otherwise anyone " +
			"who can reach it can grant themselves cluster permissions.")
	}

	return cfg
}

// Load builds a Config using getenv to read the environment variables.
func Load(getenv func(string) string) (*Config, error) {
	cfg := &Config{
		Cluster: ClusterConfig{
			Name:                getenv("CLUSTER_NAME"),
			ControlPlaneAddress: getenv("CONTROL_PLANE_ADDRESS"),
			Namespace:           getenv("NAMESPACE"),
		},
		Backend: BackendConfig{
			Port: getenv("PORT"),
		},
	}

	if cfg.Backend.Port == "" {
		return nil, fmt.Errorf("PORT env cannot be empty")
	}

	if cfg.Cluster.Name == "" {
		return nil, fmt.Errorf("CLUSTER_NAME env cannot be empty")
	}

	if cfg.Cluster.ControlPlaneAddress == "" {
		return nil, fmt.Errorf("CONTROL_PLANE_ADDRESS env cannot be empty")
	}

	if cfg.Cluster.Namespace == "" {
		return nil, fmt.Errorf("NAMESPACE env cannot be empty")
	}

	auth, err := loadAuth(getenv)
	if err != nil {
		return nil, err
	}
	cfg.Auth = auth

	kinds, err := ParseSubjectKinds(getenv("ADDITIONAL_SUBJECT_KINDS"))
	if err != nil {
		return nil, err
	}
	cfg.RBAC.AdditionalSubjectKinds = kinds

	kc, err := loadKubeconfig(getenv)
	if err != nil {
		return nil, err
	}
	cfg.Kubeconfig = kc

	return cfg, nil
}

func loadAuth(getenv func(string) string) (AuthConfig, error) {
	auth := AuthConfig{
		BasicAuthEnabled: true,
		Username:         getenv("BASIC_AUTH_USERNAME"),
		Password:         getenv("BASIC_AUTH_PASSWORD"),
	}

	if auth.Username == "" {
		auth.Username = DefaultBasicAuthUsername
	}

	if raw := strings.TrimSpace(getenv("BASIC_AUTH_DISABLED")); raw != "" {
		disabled, err := strconv.ParseBool(raw)
		if err != nil {
			// fail closed: a typo must never silently disable authentication
			return AuthConfig{}, fmt.Errorf("BASIC_AUTH_DISABLED must be a boolean (true/false), got %q", raw)
		}
		auth.BasicAuthEnabled = !disabled
	}

	if auth.BasicAuthEnabled && auth.Password == "" {
		return AuthConfig{}, fmt.Errorf("BASIC_AUTH_PASSWORD env cannot be empty (set BASIC_AUTH_DISABLED=true to disable basic auth)")
	}

	return auth, nil
}

func loadKubeconfig(getenv func(string) string) (KubeconfigConfig, error) {
	kc := KubeconfigConfig{
		TokenMode:       strings.ToLower(strings.TrimSpace(getenv("KUBECONFIG_TOKEN_MODE"))),
		TokenExpiration: DefaultTokenExpiration,
	}

	if kc.TokenMode == "" {
		kc.TokenMode = TokenModeSecret
	}

	if kc.TokenMode != TokenModeSecret && kc.TokenMode != TokenModeTokenRequest {
		return KubeconfigConfig{}, fmt.Errorf("KUBECONFIG_TOKEN_MODE must be %q or %q, got %q",
			TokenModeSecret, TokenModeTokenRequest, kc.TokenMode)
	}

	if raw := strings.TrimSpace(getenv("KUBECONFIG_TOKEN_EXPIRATION")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return KubeconfigConfig{}, fmt.Errorf("KUBECONFIG_TOKEN_EXPIRATION must be a duration (e.g. 720h): %w", err)
		}
		if d < MinTokenExpiration {
			return KubeconfigConfig{}, fmt.Errorf("KUBECONFIG_TOKEN_EXPIRATION must be at least %s, got %s", MinTokenExpiration, d)
		}
		kc.TokenExpiration = d
	}

	return kc, nil
}

// ParseSubjectKinds parses a comma separated list of RBAC subject kinds.
// Only "User" and "Group" are accepted (case-insensitive); duplicates are removed.
func ParseSubjectKinds(raw string) ([]string, error) {
	kinds := []string{}
	seen := map[string]bool{}

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var kind string
		switch strings.ToLower(part) {
		case "user":
			kind = "User"
		case "group":
			kind = "Group"
		default:
			return nil, fmt.Errorf("ADDITIONAL_SUBJECT_KINDS: unsupported subject kind %q (allowed: User, Group)", part)
		}

		if !seen[kind] {
			seen[kind] = true
			kinds = append(kinds, kind)
		}
	}

	return kinds, nil
}
