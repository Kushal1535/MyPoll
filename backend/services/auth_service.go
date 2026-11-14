package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"pulsepoll/backend/models"
)

type AuthService struct {
	collection *mongo.Collection
	jwtSecret  string
}

func NewAuthService(collection *mongo.Collection, jwtSecret string) *AuthService {
	return &AuthService{
		collection: collection,
		jwtSecret:  jwtSecret,
	}
}

// Create a unique index for user email addresses.
func (s *AuthService) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "email", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err := s.collection.Indexes().CreateOne(ctx, index)
	return err
}

// Register a new user.
func (s *AuthService) Register(
	ctx context.Context,
	name string,
	email string,
	password string,
) (*models.User, error) {

	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" || email == "" || password == "" {
		return nil, errors.New("all fields are required")
	}

	if len(password) < 8 {
		return nil, errors.New("password must contain at least 8 characters")
	}

	var existing models.User

	err := s.collection.FindOne(
		ctx,
		bson.M{"email": email},
	).Decode(&existing)

	if err == nil {
		return nil, errors.New("email is already registered")
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := models.User{
		ID:           bson.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Authenticate an existing user and generate a JWT.
func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, *models.User, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	var user models.User

	err := s.collection.FindOne(
		ctx,
		bson.M{"email": email},
	).Decode(&user)

	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)

	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.Hex(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}
