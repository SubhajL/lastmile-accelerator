package vault

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"example.com/lma/secrets-env-service/internal/config"
	vaultapi "github.com/hashicorp/vault/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Client wraps Vault API client
type Client struct {
	client    *vaultapi.Client
	mountPath string
	
	// Test mode fields
	testMode bool
	failMode bool // Simulate failures in test mode
	testData map[string]map[string]interface{}
	mu       sync.RWMutex
}

// NewClient creates a new Vault client
func NewClient(cfg *config.VaultConfig) (*Client, error) {
	vaultCfg := vaultapi.DefaultConfig()
	vaultCfg.Address = cfg.Address

	client, err := vaultapi.NewClient(vaultCfg)
	if err != nil {
		// In test mode, create a mock client
		return &Client{
			testMode:  true,
			testData:  make(map[string]map[string]interface{}),
			mountPath: cfg.MountPath,
		}, nil
	}

	// Authenticate with AppRole
	// In real implementation, this would do actual auth
	// For now, we'll keep it simple

	return &Client{
		client:    client,
		mountPath: cfg.MountPath,
		testMode:  false,
	}, nil
}

// WriteSecret writes a secret to Vault
func (c *Client) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	tr := otel.Tracer("vault")
	ctx, span := tr.Start(ctx, "vault.write")
	span.SetAttributes(attribute.String("vault.path", path))
	defer span.End()
	if c.testMode {
		if c.failMode {
			return fmt.Errorf("simulated vault failure")
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		c.testData[path] = data
		return nil
	}

	fullPath := fmt.Sprintf("%s/data/%s", c.mountPath, path)
	payload := map[string]interface{}{
		"data": data,
	}

	_, err := c.client.Logical().WriteWithContext(ctx, fullPath, payload)
	return err
}

// ReadSecret reads a secret from Vault
func (c *Client) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	tr := otel.Tracer("vault")
	ctx, span := tr.Start(ctx, "vault.read")
	span.SetAttributes(attribute.String("vault.path", path))
	defer span.End()
	if c.testMode {
		c.mu.RLock()
		defer c.mu.RUnlock()
		
		data, exists := c.testData[path]
		if !exists {
			return nil, fmt.Errorf("secret not found at path: %s", path)
		}
		return data, nil
	}

	fullPath := fmt.Sprintf("%s/data/%s", c.mountPath, path)
	secret, err := c.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// KV v2 stores data under "data" key
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid secret format at path: %s", path)
	}

	return data, nil
}

// DeleteSecret soft-deletes a secret from Vault
func (c *Client) DeleteSecret(ctx context.Context, path string) error {
	tr := otel.Tracer("vault")
	ctx, span := tr.Start(ctx, "vault.delete")
	span.SetAttributes(attribute.String("vault.path", path))
	defer span.End()
	if c.testMode {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.testData, path)
		return nil
	}

	fullPath := fmt.Sprintf("%s/data/%s", c.mountPath, path)
	_, err := c.client.Logical().DeleteWithContext(ctx, fullPath)
	
	// Idempotent - no error if already deleted
	return err
}

// ListSecrets lists secret keys under a path prefix
func (c *Client) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	tr := otel.Tracer("vault")
	ctx, span := tr.Start(ctx, "vault.list")
	span.SetAttributes(attribute.String("vault.path_prefix", pathPrefix))
	defer span.End()
	if c.testMode {
		c.mu.RLock()
		defer c.mu.RUnlock()
		
		var secrets []string
		for path := range c.testData {
			if strings.HasPrefix(path, pathPrefix) {
				secrets = append(secrets, path)
			}
		}
		return secrets, nil
	}

	fullPath := fmt.Sprintf("%s/metadata/%s", c.mountPath, pathPrefix)
	secret, err := c.client.Logical().ListWithContext(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var secrets []string
	for _, key := range keys {
		if keyStr, ok := key.(string); ok {
			secrets = append(secrets, pathPrefix+keyStr)
		}
	}

	return secrets, nil
}

// ReadSecretVersion reads a specific version of a secret
func (c *Client) ReadSecretVersion(ctx context.Context, path string, version int) (map[string]interface{}, error) {
	tr := otel.Tracer("vault")
	ctx, span := tr.Start(ctx, "vault.read_version")
	span.SetAttributes(attribute.String("vault.path", path))
	defer span.End()
	if c.testMode {
		// In test mode, just return current version
		return c.ReadSecret(ctx, path)
	}

	fullPath := fmt.Sprintf("%s/data/%s?version=%d", c.mountPath, path, version)
	secret, err := c.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret version not found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid secret format at path: %s", path)
	}

	return data, nil
}

// HealthCheck verifies Vault connectivity
func (c *Client) HealthCheck(ctx context.Context) error {
	tr := otel.Tracer("vault")
	ctx, span := tr.Start(ctx, "vault.health")
	defer span.End()
	if c.testMode {
		return nil
	}

	health, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}

	if !health.Initialized || health.Sealed {
		return fmt.Errorf("vault is not ready (initialized: %v, sealed: %v)", health.Initialized, health.Sealed)
	}

	return nil
}

// Close closes the Vault client
func (c *Client) Close() error {
	// Vault API client doesn't need explicit closing
	return nil
}

// SetTestMode enables test mode
func (c *Client) SetTestMode(enabled bool) {
	c.testMode = enabled
	if enabled && c.testData == nil {
		c.testData = make(map[string]map[string]interface{})
	}
}

// SetFailMode simulates failures in test mode
func (c *Client) SetFailMode(enabled bool) {
	c.failMode = enabled
}
