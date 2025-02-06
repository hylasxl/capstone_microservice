package handlers

import (
	"auth_service/models"
	"auth_service/proto/auth_service"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"os"
	"time"
)

var JwtSecretKey = os.Getenv("JWT_SECRET_KEY")

type AuthService struct {
	auth_service.UnimplementedAuthServiceServer
	DB *gorm.DB
}

type CustomClaims struct {
	AccountID  string   `json:"account_id"`
	Permission []string `json:"permission"`
	RoleID     string   `json:"role_id"`
	jwt.RegisteredClaims
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		DB: db,
	}
}
func (s *AuthService) ValidateToken(ctx context.Context, req *auth_service.ValidateTokenRequest) (*auth_service.ValidateTokenResponse, error) {
	tokenStr := req.Token

	// Parse the token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			println("unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JwtSecretKey), nil
	})

	// Handle token parsing errors
	if err != nil {
		println("Invalid token format or signing method")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "Invalid token format or signing method",
		}, nil
	}

	if !token.Valid {
		println("Invalid token")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "Invalid token",
		}, nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		println("Invalid token claims")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "Invalid token claims",
		}, nil
	}

	if exp, ok := claims["exp"].(float64); ok {
		expirationTime := time.Unix(int64(exp), 0)
		if time.Now().After(expirationTime) {
			println("Token expired")
			return &auth_service.ValidateTokenResponse{
				Valid:        false,
				ErrorMessage: "Token has expired",
			}, nil
		}
	} else {
		println("Missing or invalid expiration claim")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "Missing or invalid expiration claim",
		}, nil
	}

	// Validate and extract role_id
	roleID, ok := claims["role_id"].(string)
	if !ok || roleID == "" {
		println("Missing or invalid role claim")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "Missing or invalid role ID in token claims",
		}, nil
	}

	// Validate and extract user_id
	userID, ok := claims["account_id"].(string)
	if !ok || userID == "" {
		println("Missing or invalid account ID claim")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "Missing or invalid user ID in token claims",
		}, nil
	}

	// Fetch permissions for the role
	permissions, err := s.getUserPermissions(roleID)
	fmt.Printf("%v", permissions)
	if err != nil {
		println("Error getting permissions")
		return &auth_service.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("Failed to fetch permissions: %v", err),
		}, nil
	}

	return &auth_service.ValidateTokenResponse{
		Valid:       true,
		UserId:      userID,
		RoleId:      roleID,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) getUserPermissions(roleID string) ([]string, error) {
	var permissionByRoles []models.PermissionByRole
	var permissionURLs []string

	if err := s.DB.Where("role_id = ?", roleID).Find(&permissionByRoles).Error; err != nil {
		return nil, errors.New("unable to fetch user permissions")
	}

	for _, permRole := range permissionByRoles {
		var permission models.Permission
		if err := s.DB.First(&permission, permRole.Permission).Error; err != nil {
			return nil, errors.New("unable to fetch user permission details")
		}
		permissionURLs = append(permissionURLs, permission.PermissionURL)
	}

	return permissionURLs, nil
}

func (s *AuthService) GetPermissions(context context.Context, req *auth_service.GetPermissionsRequest) (*auth_service.GetPermissionsResponse, error) {
	var permissionUrls []string

	err := s.DB.
		Model(&models.PermissionByRole{}).
		Select("permissions.permission_url").
		Joins("JOIN permissions ON permissions.id = permission_by_roles.permission_id").
		Where("permission_by_roles.role_id = ?", req.RoleId).
		Pluck("permission_url", &permissionUrls).Error

	if err != nil {
		return &auth_service.GetPermissionsResponse{
				Error: "Unable to fetch permissions",
			},
			nil
	}

	return &auth_service.GetPermissionsResponse{
		Url: permissionUrls,
	}, nil
}

func (s *AuthService) GenerateTokens(context context.Context, req *auth_service.GenerateTokensRequest) (*auth_service.GenerateTokensResponse, error) {

	accessTokenENV := os.Getenv("ACCESS_TOKEN_DURATION")

	accessTokenDuration, err := time.ParseDuration(accessTokenENV)
	if err != nil {
		accessTokenDuration = time.Minute * 15
	}

	refreshTokenENV := os.Getenv("REFRESH_TOKEN_DURATION")
	refreshTokenDuration, err := time.ParseDuration(refreshTokenENV)
	if err != nil {
		refreshTokenDuration = time.Hour * 24 * 365
	}

	accessTokenClaim := CustomClaims{
		AccountID:  req.Claims.AccountId,
		Permission: req.Claims.Permissions,
		RoleID:     req.Claims.RoleId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			Issuer:    "SyncIO",
			Subject:   "Authentication",
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshTokenClaim := CustomClaims{
		AccountID: req.Claims.AccountId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenDuration)),
			Issuer:    "SyncIO",
			Subject:   "Authentication",
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		accessTokenClaim,
	)
	signedAccessToken, err := accessToken.SignedString([]byte(JwtSecretKey))

	refreshToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		refreshTokenClaim,
	)

	signedRefreshToken, err := refreshToken.SignedString([]byte(JwtSecretKey))

	return &auth_service.GenerateTokensResponse{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, nil
}
