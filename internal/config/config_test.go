package config_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"sighupio/permission-manager/internal/config"
)

func ExampleNew() {
	os.Setenv("PORT", "4000")
	os.Setenv("CLUSTER_NAME", "my-cluster")
	os.Setenv("CONTROL_PLANE_ADDRESS", "https://192.168.64.33:8443")
	os.Setenv("NAMESPACE", "test")
	os.Setenv("BASIC_AUTH_PASSWORD", "secret")

	cfg := config.New()

	fmt.Println(cfg.Backend.Port)
	fmt.Println(cfg.Cluster.Name)
	fmt.Println(cfg.Cluster.ControlPlaneAddress)
	fmt.Println(cfg.Cluster.Namespace)

	// Output:
	// 4000
	// my-cluster
	// https://192.168.64.33:8443
	// test
}

func env(overrides map[string]string) func(string) string {
	base := map[string]string{
		"PORT":                  "4000",
		"CLUSTER_NAME":          "my-cluster",
		"CONTROL_PLANE_ADDRESS": "https://k8s.example.com:6443",
		"NAMESPACE":             "permission-manager",
		"BASIC_AUTH_PASSWORD":   "secret",
	}
	for k, v := range overrides {
		base[k] = v
	}
	return func(k string) string { return base[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load(env(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Auth.BasicAuthEnabled {
		t.Error("basic auth must be enabled by default")
	}
	if cfg.Auth.Username != "admin" {
		t.Errorf("default username = %q, want admin", cfg.Auth.Username)
	}
	if cfg.Kubeconfig.TokenMode != config.TokenModeSecret {
		t.Errorf("default token mode = %q, want %q", cfg.Kubeconfig.TokenMode, config.TokenModeSecret)
	}
	if cfg.Kubeconfig.TokenExpiration != config.DefaultTokenExpiration {
		t.Errorf("default expiration = %s", cfg.Kubeconfig.TokenExpiration)
	}
	if len(cfg.RBAC.AdditionalSubjectKinds) != 0 {
		t.Errorf("no additional subject kinds expected by default, got %v", cfg.RBAC.AdditionalSubjectKinds)
	}
}

func TestLoadRequiredVariables(t *testing.T) {
	for _, key := range []string{"PORT", "CLUSTER_NAME", "CONTROL_PLANE_ADDRESS", "NAMESPACE"} {
		t.Run(key, func(t *testing.T) {
			_, err := config.Load(env(map[string]string{key: ""}))
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("expected error mentioning %s, got %v", key, err)
			}
		})
	}
}

func TestLoadBasicAuth(t *testing.T) {
	tests := []struct {
		name        string
		overrides   map[string]string
		wantErr     bool
		wantEnabled bool
		wantUser    string
	}{
		{name: "enabled with password", overrides: nil, wantEnabled: true, wantUser: "admin"},
		{name: "missing password fails closed", overrides: map[string]string{"BASIC_AUTH_PASSWORD": ""}, wantErr: true},
		{name: "explicitly disabled", overrides: map[string]string{"BASIC_AUTH_PASSWORD": "", "BASIC_AUTH_DISABLED": "true"}, wantEnabled: false, wantUser: "admin"},
		{name: "disabled=false keeps auth", overrides: map[string]string{"BASIC_AUTH_DISABLED": "false"}, wantEnabled: true, wantUser: "admin"},
		{name: "disabled=false without password fails", overrides: map[string]string{"BASIC_AUTH_DISABLED": "false", "BASIC_AUTH_PASSWORD": ""}, wantErr: true},
		{name: "invalid disabled value fails closed", overrides: map[string]string{"BASIC_AUTH_DISABLED": "yes please"}, wantErr: true},
		{name: "custom username", overrides: map[string]string{"BASIC_AUTH_USERNAME": "root"}, wantEnabled: true, wantUser: "root"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(env(tt.overrides))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Auth.BasicAuthEnabled != tt.wantEnabled {
				t.Errorf("BasicAuthEnabled = %v, want %v", cfg.Auth.BasicAuthEnabled, tt.wantEnabled)
			}
			if cfg.Auth.Username != tt.wantUser {
				t.Errorf("Username = %q, want %q", cfg.Auth.Username, tt.wantUser)
			}
		})
	}
}

func TestLoadKubeconfigSettings(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		wantErr   bool
		wantMode  string
		wantExp   time.Duration
	}{
		{name: "tokenrequest", overrides: map[string]string{"KUBECONFIG_TOKEN_MODE": "TokenRequest"}, wantMode: "tokenrequest", wantExp: config.DefaultTokenExpiration},
		{name: "custom expiration", overrides: map[string]string{"KUBECONFIG_TOKEN_MODE": "tokenrequest", "KUBECONFIG_TOKEN_EXPIRATION": "24h"}, wantMode: "tokenrequest", wantExp: 24 * time.Hour},
		{name: "unknown mode", overrides: map[string]string{"KUBECONFIG_TOKEN_MODE": "magic"}, wantErr: true},
		{name: "bad duration", overrides: map[string]string{"KUBECONFIG_TOKEN_EXPIRATION": "forever"}, wantErr: true},
		{name: "too short", overrides: map[string]string{"KUBECONFIG_TOKEN_EXPIRATION": "1m"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(env(tt.overrides))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Kubeconfig.TokenMode != tt.wantMode {
				t.Errorf("TokenMode = %q, want %q", cfg.Kubeconfig.TokenMode, tt.wantMode)
			}
			if cfg.Kubeconfig.TokenExpiration != tt.wantExp {
				t.Errorf("TokenExpiration = %s, want %s", cfg.Kubeconfig.TokenExpiration, tt.wantExp)
			}
		})
	}
}

func TestParseSubjectKinds(t *testing.T) {
	tests := []struct {
		raw     string
		want    []string
		wantErr bool
	}{
		{raw: "", want: []string{}},
		{raw: "User", want: []string{"User"}},
		{raw: " user , GROUP ", want: []string{"User", "Group"}},
		{raw: "User,user", want: []string{"User"}},
		{raw: "ServiceAccount", wantErr: true},
		{raw: "User,Robot", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := config.ParseSubjectKinds(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
