package handlers

import "strings"

// EnvAllowlist is a simple allowlist for customer "environment" values (e.g. dev/staging/prod).
//
// Zero value allows all environments.
type EnvAllowlist struct {
	set map[string]struct{}
}

func NewEnvAllowlist(values []string) EnvAllowlist {
	if len(values) == 0 {
		return EnvAllowlist{}
	}
	set := map[string]struct{}{}
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		set[strings.ToLower(v)] = struct{}{}
	}
	if len(set) == 0 {
		return EnvAllowlist{}
	}
	return EnvAllowlist{set: set}
}

// Allows checks if env is permitted. If the allowlist is configured (non-empty),
// empty env is treated as not allowed.
func (a EnvAllowlist) Allows(env string) bool {
	if len(a.set) == 0 {
		return true
	}
	env = strings.ToLower(strings.TrimSpace(env))
	if env == "" {
		return false
	}
	_, ok := a.set[env]
	return ok
}

// AllowsOptional allows the empty string (useful for query params that are optional),
// and otherwise enforces Allows().
func (a EnvAllowlist) AllowsOptional(env string) bool {
	if strings.TrimSpace(env) == "" {
		return true
	}
	return a.Allows(env)
}

