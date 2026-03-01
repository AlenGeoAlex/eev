package httpx

type contextKey string
type cookieKey string

const (
	BodyKey             contextKey = "validated_body"
	UserClaimsKey       contextKey = "user_claims"
	OAuthStateCookieKey cookieKey  = "oauth_state"

	AccessTokenCookieKey  cookieKey = "access_token"
	RefreshTokenCookieKey cookieKey = "refresh_token"
)
