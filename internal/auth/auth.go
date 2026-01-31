package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserExists         = errors.New("user already exists")
	ErrOrgExists          = errors.New("organization already exists")
)

// User represents a user in the system
type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DefaultOrgID  string    `json:"default_org_id,omitempty"`
}

// Organization represents an organization
type Organization struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	OwnerID              string     `json:"owner_id"`
	Plan                 string     `json:"plan"`
	StripeCustomerID     string     `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID string     `json:"stripe_subscription_id,omitempty"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// APIKey represents an API key for CLI authentication
type APIKey struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	OrganizationID string     `json:"organization_id"`
	Name           string     `json:"name"`
	KeyHash        string     `json:"-"`
	KeyPrefix      string     `json:"key_prefix"`
	Key            string     `json:"key,omitempty"` // Only returned on creation
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	OrgID  string `json:"org_id"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	db        *sql.DB
	jwtSecret []byte
}

// NewAuthService creates a new auth service
func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user
func (as *AuthService) Register(email, password, name string) (*User, string, error) {
	// Check if user exists
	var exists bool
	err := as.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", ErrUserExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// Create user
	user := &User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
	}

	// Insert into database
	err = as.db.QueryRow(`
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, email, string(hash), name).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, "", err
	}

	// Create default organization
	org, err := as.createDefaultOrganization(user.ID, name)
	if err != nil {
		return nil, "", err
	}

	user.DefaultOrgID = org.ID

	// Generate JWT token
	token, err := as.GenerateToken(user, org.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login authenticates a user
func (as *AuthService) Login(email, password string) (*User, string, error) {
	var user User
	err := as.db.QueryRow(`
		SELECT id, email, password_hash, name, avatar_url, email_verified, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Get user's default organization
	var orgID string
	err = as.db.QueryRow(`
		SELECT organization_id FROM organization_members
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, user.ID).Scan(&orgID)

	if err != nil && err != sql.ErrNoRows {
		return nil, "", err
	}

	user.DefaultOrgID = orgID

	// Generate JWT token
	token, err := as.GenerateToken(&user, orgID)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

// GenerateToken creates a JWT token
func (as *AuthService) GenerateToken(user *User, orgID string) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		OrgID:  orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(as.jwtSecret)
}

// VerifyToken validates a JWT token
func (as *AuthService) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return as.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GenerateAPIKey creates an API key for CLI authentication
func (as *AuthService) GenerateAPIKey(userID, orgID, name string) (*APIKey, error) {
	// Generate random key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	keyString := "ops_" + base64.URLEncoding.EncodeToString(b)

	// Hash for storage
	hash, err := bcrypt.GenerateFromPassword([]byte(keyString), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Store in database
	var apiKey APIKey
	err = as.db.QueryRow(`
		INSERT INTO api_keys (user_id, organization_id, name, key_hash, key_prefix)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, userID, orgID, name, string(hash), keyString[:12]).Scan(&apiKey.ID, &apiKey.CreatedAt)

	if err != nil {
		return nil, err
	}

	apiKey.UserID = userID
	apiKey.OrganizationID = orgID
	apiKey.Name = name
	apiKey.KeyPrefix = keyString[:12]
	apiKey.Key = keyString // Return full key only once

	return &apiKey, nil
}

// VerifyAPIKey validates an API key
func (as *AuthService) VerifyAPIKey(keyString string) (*Claims, error) {
	// Get all API keys (we need to check hash)
	rows, err := as.db.Query(`
		SELECT ak.id, ak.user_id, ak.organization_id, ak.key_hash, u.email
		FROM api_keys ak
		JOIN users u ON ak.user_id = u.id
		WHERE ak.expires_at IS NULL OR ak.expires_at > NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, userID, orgID, keyHash, email string
		if err := rows.Scan(&id, &userID, &orgID, &keyHash, &email); err != nil {
			continue
		}

		// Check if key matches
		if err := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(keyString)); err == nil {
			// Update last used
			as.db.Exec("UPDATE api_keys SET last_used_at = NOW() WHERE id = $1", id)

			return &Claims{
				UserID: userID,
				Email:  email,
				OrgID:  orgID,
			}, nil
		}
	}

	return nil, ErrInvalidToken
}

// createDefaultOrganization creates a default organization for a new user
func (as *AuthService) createDefaultOrganization(userID, name string) (*Organization, error) {
	slug := generateSlug(name)

	// Check if slug exists
	var exists bool
	err := as.db.QueryRow("SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1)", slug).Scan(&exists)
	if err != nil {
		return nil, err
	}

	// Make slug unique if needed
	if exists {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	// Create organization
	var org Organization
	trialEnds := time.Now().Add(14 * 24 * time.Hour) // 14 day trial

	err = as.db.QueryRow(`
		INSERT INTO organizations (name, slug, owner_id, plan, trial_ends_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, name+"'s Organization", slug, userID, "free", trialEnds).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return nil, err
	}

	org.Name = name + "'s Organization"
	org.Slug = slug
	org.OwnerID = userID
	org.Plan = "free"
	org.TrialEndsAt = &trialEnds

	// Add user as owner
	_, err = as.db.Exec(`
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
	`, org.ID, userID, "owner")

	if err != nil {
		return nil, err
	}

	return &org, nil
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Simple slug generation - in production use a proper library
	slug := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c + 32) // Convert to lowercase
		} else if c == ' ' || c == '-' {
			slug += "-"
		}
	}
	return slug
}
