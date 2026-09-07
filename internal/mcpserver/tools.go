package mcpserver

import (
	"context"
	"fmt"

	"github.com/ars1364/warden-mcp/internal/crypto"
	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type secretMeta struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type listSecretsOutput struct {
	Secrets []secretMeta `json:"secrets"`
}

type getSecretArgs struct {
	Name string `json:"name" jsonschema:"the secret's name"`
}

type getSecretOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type setSecretArgs struct {
	Name        string   `json:"name" jsonschema:"the secret's name"`
	Value       string   `json:"value" jsonschema:"the secret's plaintext value"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type setSecretOutput struct {
	Name string `json:"name"`
}

func hasScope(key *store.APIKey, scope string) bool {
	for _, s := range key.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func listSecretsHandler(db *store.DB, key *store.APIKey) func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, listSecretsOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listSecretsOutput, error) {
		secrets, err := db.ListSecrets()
		if err != nil {
			return nil, listSecretsOutput{}, err
		}
		out := listSecretsOutput{Secrets: []secretMeta{}}
		for _, s := range secrets {
			out.Secrets = append(out.Secrets, secretMeta{Name: s.Name, Description: s.Description, Tags: s.Tags})
		}
		_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "read", Detail: "list_secrets"})
		return nil, out, nil
	}
}

func getSecretHandler(db *store.DB, box *crypto.Box, key *store.APIKey) func(context.Context, *mcp.CallToolRequest, getSecretArgs) (*mcp.CallToolResult, getSecretOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args getSecretArgs) (*mcp.CallToolResult, getSecretOutput, error) {
		if !hasScope(key, "read") {
			return nil, getSecretOutput{}, fmt.Errorf("api key %q lacks read scope", key.Name)
		}
		id, nonce, ciphertext, err := db.GetSecretValueByName(args.Name)
		if err != nil {
			return nil, getSecretOutput{}, fmt.Errorf("secret %q not found", args.Name)
		}
		value, err := box.Open(nonce, ciphertext)
		if err != nil {
			return nil, getSecretOutput{}, err
		}
		_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "read", SecretName: args.Name, Detail: id})
		return nil, getSecretOutput{Name: args.Name, Value: value}, nil
	}
}

func setSecretHandler(db *store.DB, box *crypto.Box, key *store.APIKey) func(context.Context, *mcp.CallToolRequest, setSecretArgs) (*mcp.CallToolResult, setSecretOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args setSecretArgs) (*mcp.CallToolResult, setSecretOutput, error) {
		if !hasScope(key, "write") {
			return nil, setSecretOutput{}, fmt.Errorf("api key %q lacks write scope", key.Name)
		}
		if args.Name == "" {
			return nil, setSecretOutput{}, fmt.Errorf("name is required")
		}
		nonce, ciphertext, err := box.Seal(args.Value)
		if err != nil {
			return nil, setSecretOutput{}, err
		}

		existingID, _, _, err := db.GetSecretValueByName(args.Name)
		if err == nil {
			if updErr := db.UpdateSecret(existingID, args.Description, args.Tags, nonce, ciphertext); updErr != nil {
				return nil, setSecretOutput{}, updErr
			}
		} else {
			if _, createErr := db.CreateSecret(args.Name, args.Description, args.Tags, nonce, ciphertext, "mcp:"+key.Name); createErr != nil {
				return nil, setSecretOutput{}, createErr
			}
		}
		_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "update", SecretName: args.Name})
		return nil, setSecretOutput{Name: args.Name}, nil
	}
}
