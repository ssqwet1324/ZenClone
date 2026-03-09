package entity

import "errors"

// errors
var (
	ErrSaveRefreshToken        = errors.New("failed to save refresh token")
	ErrGetRefreshToken         = errors.New("refresh token not found in Redis and UsersService")
	ErrSignToken               = errors.New("failed to sign token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrCannotParseClaims       = errors.New("cannot parse token claims")
	ErrUserIDNotFound          = errors.New("user ID not found in token claims")
	ErrHashPassword            = errors.New("failed to hash password")
	ErrRegisterUser            = errors.New("failed to register user")
	ErrGenerateAccessToken     = errors.New("failed to generate access token")
	ErrCompareAuthData         = errors.New("failed to compare auth data")
	ErrUpdateRefreshToken      = errors.New("failed to update refresh token")
	ErrInvalidAuthHeader       = errors.New("invalid authorization header")
	ErrRefreshTokenMismatch    = errors.New("refresh token mismatch")
	ErrInternalServer          = errors.New("internal server error")
)
