package service

type AuthContext struct {
	UserID string
	Role   string
}

type SecurityConfig struct {
	Issuer   string
	Audience string
}

type AuthorizationService struct{}

func NewAuthorizationService() *AuthorizationService { return &AuthorizationService{} }

func (s *AuthorizationService) CanAccess(user AuthContext, resource string) bool {
	if user.Role == "admin" {
		return true
	}
	if user.Role == "manager" {
		return resource == "inventory" || resource == "orders"
	}
	return resource == "profile"
}

func (s *AuthorizationService) ValidateToken(claims JWTClaims) bool {
	return claims.GetSubject() != "" && claims.GetRole() != ""
}

type JWTClaims interface {
	GetSubject() string
	GetRole() string
}

type DefaultJWTClaims struct {
	Subject string
	Role    string
}

func (c DefaultJWTClaims) GetSubject() string { return c.Subject }
func (c DefaultJWTClaims) GetRole() string { return c.Role }
