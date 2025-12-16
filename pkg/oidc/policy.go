package oidc

import (
	"errors"
	"fmt"
	"strings"
)

// Policy defines rules for validating OIDC claims
type Policy struct {
	// AllowedRepositories is a list of allowed repository names (e.g., "owner/repo")
	AllowedRepositories []string

	// AllowedRepositoryOwners is a list of allowed repository owners
	AllowedRepositoryOwners []string

	// AllowedRefs is a list of allowed git refs (e.g., "refs/heads/main")
	AllowedRefs []string

	// AllowedActors is a list of allowed GitHub usernames
	AllowedActors []string

	// AllowedWorkflows is a list of allowed workflow paths
	AllowedWorkflows []string
}

// Check validates the claims against the policy
func (p *Policy) Check(claims *Claims) error {
	if claims == nil {
		return errors.New("claims cannot be nil")
	}

	// Check repository
	if len(p.AllowedRepositories) > 0 {
		if !contains(p.AllowedRepositories, claims.Repository) {
			return fmt.Errorf("repository %q is not allowed", claims.Repository)
		}
	}

	// Check repository owner
	if len(p.AllowedRepositoryOwners) > 0 {
		if !contains(p.AllowedRepositoryOwners, claims.RepositoryOwner) {
			return fmt.Errorf("repository owner %q is not allowed", claims.RepositoryOwner)
		}
	}

	// Check ref
	if len(p.AllowedRefs) > 0 {
		if !containsPrefix(p.AllowedRefs, claims.Ref) {
			return fmt.Errorf("ref %q is not allowed", claims.Ref)
		}
	}

	// Check actor
	if len(p.AllowedActors) > 0 {
		if !contains(p.AllowedActors, claims.Actor) {
			return fmt.Errorf("actor %q is not allowed", claims.Actor)
		}
	}

	// Check workflow
	if len(p.AllowedWorkflows) > 0 {
		if !containsSuffix(p.AllowedWorkflows, claims.JobWorkflowRef) {
			return fmt.Errorf("workflow %q is not allowed", claims.JobWorkflowRef)
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsPrefix(prefixes []string, item string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}
	return false
}

func containsSuffix(suffixes []string, item string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(item, suffix) {
			return true
		}
	}
	return false
}
