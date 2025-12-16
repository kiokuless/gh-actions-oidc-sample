package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type AuthRequest struct {
	Token string `json:"token"`
}

type AuthResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message,omitempty"`
	Claims      json.RawMessage `json:"claims,omitempty"`
	AccessToken string          `json:"access_token,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "auth":
		authCmd(os.Args[2:])
	case "health":
		healthCmd(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gh-actions-oidc-cli - CLI tool for OIDC authentication

Usage:
  gh-actions-oidc-cli <command> [options]

Commands:
  auth    Authenticate with OIDC token
  health  Check server health
  help    Show this help message

Examples:
  # Check server health
  gh-actions-oidc-cli health -server http://localhost:8080

  # Authenticate with OIDC token (from GitHub Actions)
  gh-actions-oidc-cli auth -server http://localhost:8080 -token $OIDC_TOKEN

  # Authenticate using token from environment
  OIDC_TOKEN=xxx gh-actions-oidc-cli auth -server http://localhost:8080`)
}

func authCmd(args []string) {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "OIDC server URL")
	token := fs.String("token", "", "OIDC token (or use OIDC_TOKEN env var)")
	fs.Parse(args)

	if *token == "" {
		*token = os.Getenv("OIDC_TOKEN")
	}
	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: token is required (-token flag or OIDC_TOKEN env var)")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	reqBody, _ := json.Marshal(AuthRequest{Token: *token})
	resp, err := client.Post(*serverURL+"/auth", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !authResp.Success {
		fmt.Fprintf(os.Stderr, "Authentication failed: %s\n", authResp.Message)
		os.Exit(1)
	}

	fmt.Println("Authentication successful!")
	fmt.Printf("Access Token: %s\n", authResp.AccessToken)

	if len(authResp.Claims) > 0 {
		fmt.Println("\nClaims:")
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, authResp.Claims, "", "  ")
		fmt.Println(prettyJSON.String())
	}
}

func healthCmd(args []string) {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "OIDC server URL")
	fs.Parse(args)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(*serverURL + "/health")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server health check failed: status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("Server is healthy")
}
