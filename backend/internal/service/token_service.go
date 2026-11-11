package service

type TokenService interface {
    Generate(userID string) (string, error)
    Verify(token string) (*Claims, error)
}
type Claims struct {
	UserID string `json:"user_id"`
}
type JWTTokenService struct {
    
}
