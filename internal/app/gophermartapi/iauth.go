package gophermartapi

type Auth interface {
	GenerateToken(userID int64) (string, error)
	ParseToken(accessToken string) (int64, error)
	GeneratePasswordHash(password string, salt string) string
	GenerateUserSalt(complexity int64) string
}
