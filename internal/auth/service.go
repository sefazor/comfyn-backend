package auth

import (
	"errors"

	"github.com/sefazor/comfyn/internal/models"
	"github.com/sefazor/comfyn/pkg/database"
	"github.com/sefazor/comfyn/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

func Register(input RegisterInput) (*AuthResponse, error) {
	// Email ve username kontrolü
	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?", input.Email, input.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("email or username already exists")
	}

	// Şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Yeni kullanıcı oluştur
	user := models.User{
		FullName:     input.FullName,
		Email:        input.Email,
		Username:     input.Username,
		Password:     string(hashedPassword),
		ProfileImage: "https://example.com/default-profile.jpg",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	// JWT token oluştur
	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user.SafeResponse(),
	}, nil
}

func Login(input LoginInput) (*AuthResponse, error) {
	var user models.User

	// Kullanıcıyı bul (username veya email ile)
	if err := database.DB.Where("username = ? OR email = ?", input.Username, input.Username).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Şifreyi kontrol et
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// JWT token oluştur
	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user.SafeResponse(),
	}, nil
}
