package oidc

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	// GitHubOIDCIssuer is the OIDC issuer URL for GitHub Actions
	GitHubOIDCIssuer = "https://token.actions.githubusercontent.com"
)

// Claims represents the GitHub Actions OIDC token claims
type Claims struct {
	// Standard OIDC claims
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`

	// GitHub-specific claims
	Repository      string `json:"repository"`
	RepositoryOwner string `json:"repository_owner"`
	JobWorkflowRef  string `json:"job_workflow_ref"`
	Ref             string `json:"ref"`
	RefType         string `json:"ref_type"`
	RunID           string `json:"run_id"`
	RunNumber       string `json:"run_number"`
	Actor           string `json:"actor"`
	Workflow        string `json:"workflow"`
	EventName       string `json:"event_name"`
}

// Verifier verifies GitHub Actions OIDC tokens
type Verifier struct {
	provider *oidc.Provider
	audience string
}

// VerifierConfig holds configuration for creating a Verifier
type VerifierConfig struct {
	Audience string
}

// NewVerifier creates a new OIDC verifier for GitHub Actions
func NewVerifier(ctx context.Context, cfg VerifierConfig) (*Verifier, error) {
	if cfg.Audience == "" {
		return nil, errors.New("audience is required")
	}

	provider, err := oidc.NewProvider(ctx, GitHubOIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	return &Verifier{
		provider: provider,
		audience: cfg.Audience,
	}, nil
}

// Verify verifies the OIDC token and returns the claims
func (v *Verifier) Verify(ctx context.Context, rawToken string) (*Claims, error) {
	verifier := v.provider.Verifier(&oidc.Config{
		ClientID: v.audience,
	})

	token, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}

// VerifyWithPolicy verifies the token and checks against a policy
func (v *Verifier) VerifyWithPolicy(ctx context.Context, rawToken string, policy *Policy) (*Claims, error) {
	claims, err := v.Verify(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	if err := policy.Check(claims); err != nil {
		return nil, fmt.Errorf("policy check failed: %w", err)
	}

	return claims, nil
}
