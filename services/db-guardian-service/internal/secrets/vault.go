package secrets

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

type VaultConfig struct {
	Address  string
	RoleID   string
	SecretID string
}

type VaultClient struct {
	client *vault.Client
}

func extractProjectDSN(secret map[string]string) (string, error) {
	for _, key := range []string{"dsn", "DSN", "database_url", "DATABASE_URL", "url", "URL"} {
		if v := secret[key]; v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("dsn not found in secret")
}

func NewVaultClient(cfg *VaultConfig) (*VaultClient, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("Vault address is required")
	}
	if cfg.RoleID == "" {
		return nil, fmt.Errorf("Vault RoleID is required")
	}
	if cfg.SecretID == "" {
		return nil, fmt.Errorf("Vault SecretID is required")
	}

	config := vault.DefaultConfig()
	config.Address = cfg.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Authenticate using AppRole
	data := map[string]interface{}{
		"role_id":   cfg.RoleID,
		"secret_id": cfg.SecretID,
	}

	resp, err := client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Vault: %w", err)
	}

	if resp == nil || resp.Auth == nil {
		return nil, fmt.Errorf("authentication response from Vault is invalid")
	}

	client.SetToken(resp.Auth.ClientToken)

	return &VaultClient{client: client}, nil
}

func (vc *VaultClient) GetProjectDSN(ctx context.Context, path string) (string, error) {
	secret, err := GetSecret(ctx, vc, path)
	if err != nil {
		return "", err
	}
	return extractProjectDSN(secret)
}

// Health checks Vault sys health; returns error if unhealthy or unreachable.
func (vc *VaultClient) Health(ctx context.Context) error {
	if vc == nil || vc.client == nil {
		return fmt.Errorf("vault client not initialized")
	}
	// The API does not support context here; call and rely on handler timeout.
	if _, err := vc.client.Sys().Health(); err != nil {
		return fmt.Errorf("vault health: %w", err)
	}
	return nil
}

func GetSecret(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
	if vc == nil {
		return nil, fmt.Errorf("Vault client is nil")
	}
	if path == "" {
		return nil, fmt.Errorf("secret path is required")
	}

	secret, err := vc.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// Handle KV v2 structure (data wrapper)
	var data map[string]interface{}
	if dataVal, ok := secret.Data["data"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			data = dataMap
		} else {
			data = secret.Data
		}
	} else {
		data = secret.Data
	}

	// Convert to map[string]string
	result := make(map[string]string)
	for k, v := range data {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}

	return result, nil
}
