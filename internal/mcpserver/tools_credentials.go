package mcpserver

import (
	"context"
	"fmt"

	"github.com/ars1364/warden-mcp/internal/authn"
	"github.com/ars1364/warden-mcp/internal/crypto"
	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getTOTPCodeArgs struct {
	Name string `json:"name" jsonschema:"the totp secret's name"`
}

type getTOTPCodeOutput struct {
	Name             string `json:"name"`
	Code             string `json:"code"`
	SecondsRemaining int    `json:"seconds_remaining"`
}

// getTOTPCodeHandler computes the current 6-digit code for a vaulted totp seed without ever
// putting the seed itself in the tool result — the convenience path for routine 2FA logins.
func getTOTPCodeHandler(db *store.DB, box *crypto.Box, key *store.APIKey) func(context.Context, *mcp.CallToolRequest, getTOTPCodeArgs) (*mcp.CallToolResult, getTOTPCodeOutput, error) {
	return func(_ context.Context, req *mcp.CallToolRequest, args getTOTPCodeArgs) (*mcp.CallToolResult, getTOTPCodeOutput, error) {
		if !hasScope(key, "read") {
			return nil, getTOTPCodeOutput{}, fmt.Errorf("api key %q lacks read scope", key.Name)
		}
		meta, err := db.GetSecretMetaByName(args.Name)
		if err != nil {
			return nil, getTOTPCodeOutput{}, fmt.Errorf("secret %q not found", args.Name)
		}
		if err := checkResourceAccess(db, key, store.ResourceSecret, args.Name, false); err != nil {
			return nil, getTOTPCodeOutput{}, err
		}
		if meta.Type != store.TypeTOTP {
			return nil, getTOTPCodeOutput{}, fmt.Errorf("secret %q is not a totp credential", args.Name)
		}

		fields, err := db.GetSecretFields(meta.ID)
		if err != nil {
			return nil, getTOTPCodeOutput{}, err
		}
		var seed string
		for _, f := range fields {
			if f.Key == "seed" {
				seed, err = box.Open(f.Nonce, f.Ciphertext)
				if err != nil {
					return nil, getTOTPCodeOutput{}, err
				}
				break
			}
		}
		if seed == "" {
			return nil, getTOTPCodeOutput{}, fmt.Errorf("secret %q has no seed field", args.Name)
		}

		code, secondsRemaining, err := authn.CurrentTOTPCode(seed)
		if err != nil {
			return nil, getTOTPCodeOutput{}, err
		}
		_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "read", SecretName: args.Name, Detail: "totp_code", IP: requestIP(req)})
		return nil, getTOTPCodeOutput{Name: args.Name, Code: code, SecondsRemaining: secondsRemaining}, nil
	}
}

type setCredentialArgs struct {
	Name        string            `json:"name" jsonschema:"the secret's name"`
	Type        string            `json:"type" jsonschema:"structured, totp, or reference"`
	Fields      map[string]string `json:"fields" jsonschema:"key/value pairs; a totp credential requires a 'seed' key"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	ExpiresAt   string            `json:"expires_at,omitempty" jsonschema:"optional YYYY-MM-DD expiry date"`
}

type setCredentialOutput struct {
	Name string `json:"name"`
}

// setCredentialHandler upserts a structured/totp/reference secret. set_secret stays the
// single-value (opaque) tool; this is its multi-field counterpart.
func setCredentialHandler(db *store.DB, box *crypto.Box, key *store.APIKey) func(context.Context, *mcp.CallToolRequest, setCredentialArgs) (*mcp.CallToolResult, setCredentialOutput, error) {
	return func(_ context.Context, req *mcp.CallToolRequest, args setCredentialArgs) (*mcp.CallToolResult, setCredentialOutput, error) {
		if !hasScope(key, "write") {
			return nil, setCredentialOutput{}, fmt.Errorf("api key %q lacks write scope", key.Name)
		}
		if args.Name == "" {
			return nil, setCredentialOutput{}, fmt.Errorf("name is required")
		}
		if err := checkResourceAccess(db, key, store.ResourceSecret, args.Name, true); err != nil {
			return nil, setCredentialOutput{}, err
		}
		secretType := args.Type
		if secretType != store.TypeStructured && secretType != store.TypeTOTP && secretType != store.TypeReference {
			return nil, setCredentialOutput{}, fmt.Errorf("type must be structured, totp, or reference")
		}
		if secretType == store.TypeTOTP {
			if _, ok := args.Fields["seed"]; !ok || args.Fields["seed"] == "" {
				return nil, setCredentialOutput{}, fmt.Errorf("totp credentials require a non-empty 'seed' field")
			}
		}
		expiresAt, err := parseExpiresAt(args.ExpiresAt)
		if err != nil {
			return nil, setCredentialOutput{}, err
		}

		sealed := make([]store.FieldCiphertext, 0, len(args.Fields))
		i := 0
		for k, v := range args.Fields {
			nonce, ciphertext, err := box.Seal(v)
			if err != nil {
				return nil, setCredentialOutput{}, err
			}
			sealed = append(sealed, store.FieldCiphertext{Key: k, Position: i, Nonce: nonce, Ciphertext: ciphertext})
			i++
		}

		meta, err := db.GetSecretMetaByName(args.Name)
		if err != nil {
			meta, err = db.CreateSecretMeta(args.Name, args.Description, args.Tags, secretType, expiresAt, "mcp:"+key.Name)
			if err != nil {
				return nil, setCredentialOutput{}, err
			}
		} else {
			if updErr := db.UpdateSecretMeta(meta.ID, args.Description, args.Tags, expiresAt); updErr != nil {
				return nil, setCredentialOutput{}, updErr
			}
		}
		if err := db.ReplaceSecretFields(meta.ID, sealed); err != nil {
			return nil, setCredentialOutput{}, err
		}

		_ = db.WriteAudit(store.AuditEntry{ActorType: "mcp_key", ActorID: key.ID, ActorLabel: key.Name, Action: "update", SecretName: args.Name, IP: requestIP(req)})
		return nil, setCredentialOutput{Name: args.Name}, nil
	}
}
