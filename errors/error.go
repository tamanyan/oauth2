package errors

import "errors"

// known errors
var (
	ErrInvalidRedirectURI   = errors.New("invalid_redirect_uri")
	ErrInvalidAuthorizeCode = errors.New("invalid_authorize_code")
	ErrInvalidAccessToken   = errors.New("invalid_access_token")
	ErrInvalidRefreshToken  = errors.New("invalid_refresh_token")
	ErrExpiredAccessToken   = errors.New("expired_access_token")
	ErrExpiredRefreshToken  = errors.New("expired_refresh_token")
)
