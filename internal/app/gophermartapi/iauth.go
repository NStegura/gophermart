package gophermartapi

type Auth interface {
	GenerateToken(userID int64) (string, error)
	ParseToken(accessToken string) (int64, error)
	GeneratePasswordHash(password string, complexity int) (string, error)
	CheckPasswordHash(password, hash string) bool
}
