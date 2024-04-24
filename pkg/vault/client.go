package vault

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/spf13/viper"
)

type Secrets map[string]any

type Client interface {
	// GetSecrets returns a map of secrets for the given path.
	GetSecrets(path string) (Secrets, error)
}

type client struct {
	v *vault.Client
}

func NewClient(vaultAddr string) (Client, error) {
	config := vault.DefaultConfig()
	config.Address = vaultAddr

	c, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	clientImpl := &client{
		v: c,
	}

	err = clientImpl.login()
	if err != nil {
		return nil, fmt.Errorf("unable to login to Vault: %w", err)
	}

	return clientImpl, nil
}

func (c *client) login() error {
	vip := viper.New()
	err := vip.BindEnv("vault.approle_id", "VAULT_APPROLE_ID")
	if err != nil {
		return fmt.Errorf("unable to bind environment variable: %w", err)
	}

	approleSecretID := &approle.SecretID{
		FromEnv: "VAULT_APPROLE_SECRET_ID",
	}

	// Authenticate with Vault with the AppRole auth method
	appRoleAuth, err := approle.NewAppRoleAuth(
		vip.GetString("vault.approle_id"),
		approleSecretID,
	)
	if err != nil {
		return fmt.Errorf("unable to create AppRole auth: %w", err)
	}

	authInfo, err := c.v.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return fmt.Errorf("unable to authenticate with Vault: %w", err)
	}
	if authInfo == nil {
		return errors.New("authentication with Vault failed")
	}

	return nil
}

func (c *client) GetSecrets(path string) (Secrets, error) {
	secret, err := c.v.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secrets: %w", err)
	}

	return secret.Data, nil
}
