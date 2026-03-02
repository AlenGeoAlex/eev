package services

import (
	sqlite_eev "backend-go/internal/db/sqlite/generated"
	"context"
	"errors"
	"time"

	"backend-go/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type GoogleUser struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
}

type AuthService struct {
	oauthConfig *oauth2.Config
	jwtConfig   config.JwtConfig
	queries     *sqlite_eev.Queries
}

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

func NewAuthService(cfg config.GoogleOAuthConfig, jwt config.JwtConfig, queries *sqlite_eev.Queries) *AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthService{oauthConfig: oauthConfig, jwtConfig: jwt, queries: queries}
}

// GetAuthURL returns the URL to redirect the user to for Google login
func (s AuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ValidateAndGetOAuthUser exchanges the code from Google's callback and returns the user
func (s AuthService) ValidateAndGetOAuthUser(ctx context.Context, code string) (*GoogleUser, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, errors.New("failed to exchange token: " + err.Error())
	}

	httpClient := s.oauthConfig.Client(ctx, token)

	svc, err := googleoauth2.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, errors.New("failed to create oauth2 service: " + err.Error())
	}

	info, err := svc.Userinfo.Get().Do()
	if err != nil {
		return nil, errors.New("failed to get user info: " + err.Error())
	}

	return &GoogleUser{
		ID:        info.Id,
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.Picture,
	}, nil
}
func (s AuthService) GenerateTokenPair(
	userID uuid.UUID,
	email string,
) (access string, refresh string, expiry time.Time, err error) {

	now := time.Now()
	pairJTI := uuid.New()
	refreshExpiry := now.Add(s.jwtConfig.RefreshTTL)

	accessClaims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"jti":   pairJTI.String(),
		"iss":   s.jwtConfig.Issuer,
		"iat":   now.Unix(),
		"exp":   now.Add(s.jwtConfig.AccessTTL).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	signedAccess, err := accessToken.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshClaims := jwt.MapClaims{
		"sub": userID.String(),
		"jti": pairJTI.String(), // SAME JTI
		"iss": s.jwtConfig.Issuer,
		"iat": now.Unix(),
		"exp": refreshExpiry.Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	signedRefresh, err := refreshToken.SignedString([]byte(s.jwtConfig.RefreshSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	err = s.queries.InsertRefreshToken(context.Background(), sqlite_eev.InsertRefreshTokenParams{
		Jti:       pairJTI.String(),
		UserID:    userID.String(),
		ExpiresAt: refreshExpiry,
	})

	if err != nil {
		return "", "", time.Time{}, err
	}

	return signedAccess, signedRefresh, refreshExpiry, nil
}

func (s AuthService) ValidateAccessToken(
	tokenStr string,
	validateExpiry bool,
) (*jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if validateExpiry {
				return nil, ErrTokenExpired
			}
		} else {
			return nil, ErrTokenInvalid
		}
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return &claims, nil
}

func (s AuthService) Refresh(
	ctx context.Context,
	accessTokenStr string,
	refreshTokenStr string,
) (access string, refresh string, expiry time.Time, err error) {
	accessToken, _, err := new(jwt.Parser).ParseUnverified(accessTokenStr, jwt.MapClaims{})
	if err != nil {
		return "", "", time.Time{}, errors.New("Invalid access token")
	}

	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessJTI := accessClaims["jti"].(string)
	sub := accessClaims["sub"].(string)
	email := accessClaims["email"].(string)

	refreshToken, err := jwt.Parse(refreshTokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtConfig.RefreshSecret), nil
	})
	if err != nil || !refreshToken.Valid {
		return "", "", time.Time{}, errors.New("invalid refresh token")
	}

	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshJTI := refreshClaims["jti"].(string)

	if accessJTI != refreshJTI {
		return "", "", time.Time{}, errors.New("token pair mismatch")
	}

	dbToken, err := s.queries.GetRefreshToken(ctx, refreshJTI)
	if err != nil {
		return "", "", time.Time{}, errors.New("refresh token not found (replay detected)")
	}

	if time.Now().After(dbToken.ExpiresAt) {
		_ = s.queries.RevokeRefreshToken(ctx, refreshJTI)
		return "", "", time.Time{}, errors.New("refresh token expired")
	}

	err = s.queries.RevokeRefreshToken(ctx, refreshJTI)
	if err != nil {
		return "", "", time.Time{}, err
	}

	userID, _ := uuid.Parse(sub)

	return s.GenerateTokenPair(userID, email)
}
