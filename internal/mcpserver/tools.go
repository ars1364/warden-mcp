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
	Type        string   `json:"type"`
}

type listSecretsOutput struct {
	Secrets []secretMeta `json:"secrets"`
}

type getSecretArgs struct {
	Name string `json:"name" jsonschema:"the secret's name"`
}

// getSecretOutput covers every secret type: Value is set for opaque secrets, Fields for
// structured/totp/reference ones. A totp secret's Fields includes the raw seed here — use
// get_totp_code instead when only the current code is needed, so the seed stays in the vault.
type getSecretOutput struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Value  string            `json:"value,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
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
			out.Secrets = append(out.Secrets, secretMeta{Name: s.Name, Description: s.Description, Tags: s.Tags, Type: s.Type})
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
		meta, err := db.GetSecretMetaByName(args.Name)
		if err != nil {
			return nil, getSecretOutput{}, fmt.Errorf("secret %q not found", args.Name)
		}

		if meta.Type == store.TypeOpaque {
			_, nonce, ciphertext, err := db.GetSecretValueByName(args.Name)
			if err != nil {
				return nil, getSecretOutput{}, err
			}
			value, err := box.Open(nonce, ciphertext)
			if err != nil {
				return nil, getSecretOutput{}, err
			}
			_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "read", SecretName: args.Name, Detail: meta.ID})
			return nil, getSecretOutput{Name: args.Name, Type: meta.Type, Value: value}, nil
		}

		fields, err := db.GetSecretFields(meta.ID)
		if err != nil {
			return nil, getSecretOutput{}, err
		}
		values := make(map[string]string, len(fields))
		for _, f := range fields {
			v, err := box.Open(f.Nonce, f.Ciphertext)
			if err != nil {
				return nil, getSecretOutput{}, err
			}
			values[f.Key] = v
		}
		_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "read", SecretName: args.Name, Detail: meta.ID})
		return nil, getSecretOutput{Name: args.Name, Type: meta.Type, Fields: values}, nil
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
