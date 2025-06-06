package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"component-4/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo contiene los datos que nos interesan de la API de Google.
type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

var googleOAuthConfig *oauth2.Config

// ConfigureGoogleOauth inicializa la configuración de OAuth2.
func ConfigureGoogleOauth(cfg *config.Config) {
	googleOAuthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClient,
		ClientSecret: cfg.GoogleSecret,
		RedirectURL:  cfg.GoogleRedirect,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GetGoogleLoginURL genera la URL a la que se debe redirigir al usuario.
func GetGoogleLoginURL() string {
	return googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// GetGoogleUserInfo intercambia el código de autorización por la información del usuario de Google.
func GetGoogleUserInfo(code string) (*GoogleUserInfo, error) {
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info, status: %s", resp.Status)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}
