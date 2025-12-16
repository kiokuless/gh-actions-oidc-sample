package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kiokuless/gh-actions-oidc-sample/pkg/oidc"
)

type Config struct {
	Addr                    string
	Audience                string
	AllowedRepositories     []string
	AllowedRepositoryOwners []string
}

type Server struct {
	verifier *oidc.Verifier
	policy   *oidc.Policy
}

type AuthRequest struct {
	Token string `json:"token"`
}

type AuthResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message,omitempty"`
	Claims      *oidc.Claims `json:"claims,omitempty"`
	AccessToken string       `json:"access_token,omitempty"`
}

func main() {
	cfg := parseFlags()

	ctx := context.Background()

	verifier, err := oidc.NewVerifier(ctx, oidc.VerifierConfig{
		Audience: cfg.Audience,
	})
	if err != nil {
		log.Fatalf("Failed to create OIDC verifier: %v", err)
	}

	policy := &oidc.Policy{
		AllowedRepositories:     cfg.AllowedRepositories,
		AllowedRepositoryOwners: cfg.AllowedRepositoryOwners,
	}

	server := &Server{
		verifier: verifier,
		policy:   policy,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/auth", server.handleAuth)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "addr", ":8080", "Server address")
	flag.StringVar(&cfg.Audience, "audience", "", "Expected OIDC audience")

	var repos, owners string
	flag.StringVar(&repos, "allowed-repos", "", "Comma-separated list of allowed repositories")
	flag.StringVar(&owners, "allowed-owners", "", "Comma-separated list of allowed repository owners")

	flag.Parse()

	if cfg.Audience == "" {
		cfg.Audience = os.Getenv("OIDC_AUDIENCE")
	}
	if cfg.Audience == "" {
		log.Fatal("audience is required (use -audience flag or OIDC_AUDIENCE env var)")
	}

	if repos == "" {
		repos = os.Getenv("ALLOWED_REPOS")
	}
	if repos != "" {
		cfg.AllowedRepositories = strings.Split(repos, ",")
	}

	if owners == "" {
		owners = os.Getenv("ALLOWED_OWNERS")
	}
	if owners != "" {
		cfg.AllowedRepositoryOwners = strings.Split(owners, ",")
	}

	return cfg
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Token == "" {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Token is required",
		})
		return
	}

	claims, err := s.verifier.VerifyWithPolicy(r.Context(), req.Token, s.policy)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Authentication failed: " + err.Error(),
		})
		return
	}

	log.Printf("Authentication successful for repository: %s, actor: %s", claims.Repository, claims.Actor)

	// Generate a simple access token (in production, use a proper token generation)
	accessToken := generateAccessToken(claims)

	respondJSON(w, http.StatusOK, AuthResponse{
		Success:     true,
		Message:     "Authentication successful",
		Claims:      claims,
		AccessToken: accessToken,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func generateAccessToken(claims *oidc.Claims) string {
	// In production, generate a proper JWT or session token
	// This is a placeholder for demonstration
	return "gha_" + claims.Repository + "_" + claims.RunID
}
