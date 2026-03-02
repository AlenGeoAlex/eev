package internal

type CookieKey string

const (
	OAuthStateCookieKey CookieKey = "oauth_state"
)

type ErrorResponse struct {
	Message string `json:"message"`
}
