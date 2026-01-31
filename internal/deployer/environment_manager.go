package deployer
import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)
type EnvironmentType string
const (
	EnvironmentProduction  EnvironmentType = "production"
	EnvironmentStaging     EnvironmentType = "staging"
	EnvironmentDevelopment EnvironmentType = "development"
	EnvironmentPreview     EnvironmentType = "preview"
	EnvironmentTesting     EnvironmentType = "testing"
	EnvironmentDemo        EnvironmentType = "demo"
	EnvironmentCustom      EnvironmentType = "custom"
)
type Environment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        EnvironmentType        `json:"type"`
	ProjectID   string                 `json:"project_id"`
	Variables   map[string]string      `json:"variables"`
	Secrets     map[string]string      `json:"secrets"`
	Domains     []string               `json:"domains"`
	Resources   ResourceAllocation     `json:"resources"`
	Locked      bool                   `json:"locked"`
	LockedBy    string                 `json:"locked_by,omitempty"`
	LockedAt    *time.Time             `json:"locked_at,omitempty"`
	AccessRoles map[string][]string    `json:"access_roles"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}
type ResourceAllocation struct {
	MinCPU      string `json:"min_cpu"`
	MaxCPU      string `json:"max_cpu"`
	MinMemory   string `json:"min_memory"`
	MaxMemory   string `json:"max_memory"`
	MinReplicas int    `json:"min_replicas"`
	MaxReplicas int    `json:"max_replicas"`
	StorageSize string `json:"storage_size"`
	AutoScale   bool   `json:"auto_scale"`
}
type EnvironmentManager struct {
	encryptionKey []byte
	storagePath   string
}
func NewEnvironmentManager(encryptionKey string, storagePath string) (*EnvironmentManager, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}
	return &EnvironmentManager{
		encryptionKey: []byte(encryptionKey),
		storagePath:   storagePath,
	}, nil
}
func (em *EnvironmentManager) CreateEnvironment(ctx context.Context, env *Environment) error {
	if env.ID == "" {
		env.ID = generateID()
	}
	env.CreatedAt = time.Now()
	env.UpdatedAt = time.Now()
	if err := em.encryptSecrets(env); err != nil {
		return fmt.Errorf("failed to encrypt secrets: %w", err)
	}
	if env.Resources.MinCPU == "" {
		em.setDefaultResources(env)
	}
	return em.saveEnvironment(env)
}
func (em *EnvironmentManager) GetEnvironment(ctx context.Context, envID string) (*Environment, error) {
	env, err := em.loadEnvironment(envID)
	if err != nil {
		return nil, err
	}
	if err := em.decryptSecrets(env); err != nil {
		return nil, fmt.Errorf("failed to decrypt secrets: %w", err)
	}
	return env, nil
}
func (em *EnvironmentManager) UpdateEnvironment(ctx context.Context, env *Environment) error {
	if env.Locked {
		return fmt.Errorf("environment is locked by %s", env.LockedBy)
	}
	env.UpdatedAt = time.Now()
	if err := em.encryptSecrets(env); err != nil {
		return fmt.Errorf("failed to encrypt secrets: %w", err)
	}
	return em.saveEnvironment(env)
}
func (em *EnvironmentManager) DeleteEnvironment(ctx context.Context, envID string) error {
	env, err := em.loadEnvironment(envID)
	if err != nil {
		return err
	}
	if env.Locked {
		return fmt.Errorf("cannot delete locked environment")
	}
	envPath := filepath.Join(em.storagePath, envID+".json")
	return os.Remove(envPath)
}
func (em *EnvironmentManager) CloneEnvironment(ctx context.Context, sourceID, targetName string, targetType EnvironmentType) (*Environment, error) {
	source, err := em.GetEnvironment(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	clone := &Environment{
		ID:          generateID(),
		Name:        targetName,
		Type:        targetType,
		ProjectID:   source.ProjectID,
		Variables:   make(map[string]string),
		Secrets:     make(map[string]string),
		Domains:     []string{},
		Resources:   source.Resources,
		AccessRoles: make(map[string][]string),
		Metadata:    make(map[string]interface{}),
	}
	for k, v := range source.Variables {
		clone.Variables[k] = v
	}
	if targetType == EnvironmentProduction || source.Type != EnvironmentProduction {
		for k, v := range source.Secrets {
			clone.Secrets[k] = v
		}
	}
	if targetType != EnvironmentProduction {
		clone.Resources.MinReplicas = 1
		clone.Resources.MaxReplicas = 2
	}
	if err := em.CreateEnvironment(ctx, clone); err != nil {
		return nil, err
	}
	return clone, nil
}
func (em *EnvironmentManager) PromoteEnvironment(ctx context.Context, sourceID, targetID string) error {
	source, err := em.GetEnvironment(ctx, sourceID)
	if err != nil {
		return err
	}
	target, err := em.GetEnvironment(ctx, targetID)
	if err != nil {
		return err
	}
	if target.Locked {
		return fmt.Errorf("target environment is locked")
	}
	excludeKeys := map[string]bool{
		"DATABASE_URL": true,
		"REDIS_URL":    true,
		"API_URL":      true,
	}
	for k, v := range source.Variables {
		if !excludeKeys[k] {
			target.Variables[k] = v
		}
	}
	return em.UpdateEnvironment(ctx, target)
}
func (em *EnvironmentManager) LockEnvironment(ctx context.Context, envID, userID string) error {
	env, err := em.loadEnvironment(envID)
	if err != nil {
		return err
	}
	if env.Locked {
		return fmt.Errorf("environment already locked by %s", env.LockedBy)
	}
	now := time.Now()
	env.Locked = true
	env.LockedBy = userID
	env.LockedAt = &now
	return em.saveEnvironment(env)
}
func (em *EnvironmentManager) UnlockEnvironment(ctx context.Context, envID, userID string) error {
	env, err := em.loadEnvironment(envID)
	if err != nil {
		return err
	}
	if !env.Locked {
		return errors.New("environment is not locked")
	}
	if env.LockedBy != userID {
		return fmt.Errorf("environment locked by different user: %s", env.LockedBy)
	}
	env.Locked = false
	env.LockedBy = ""
	env.LockedAt = nil
	return em.saveEnvironment(env)
}
func (em *EnvironmentManager) SetSecret(ctx context.Context, envID, key, value string) error {
	env, err := em.GetEnvironment(ctx, envID)
	if err != nil {
		return err
	}
	env.Secrets[key] = value
	return em.UpdateEnvironment(ctx, env)
}
func (em *EnvironmentManager) GetSecret(ctx context.Context, envID, key string) (string, error) {
	env, err := em.GetEnvironment(ctx, envID)
	if err != nil {
		return "", err
	}
	value, ok := env.Secrets[key]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", key)
	}
	return value, nil
}
func (em *EnvironmentManager) ListEnvironments(ctx context.Context, projectID string) ([]*Environment, error) {
	var environments []*Environment
	files, err := os.ReadDir(em.storagePath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		envID := file.Name()[:len(file.Name())-5]
		env, err := em.loadEnvironment(envID)
		if err != nil {
			continue
		}
		if env.ProjectID == projectID {
			if err := em.decryptSecrets(env); err != nil {
				continue
			}
			environments = append(environments, env)
		}
	}
	return environments, nil
}
func (em *EnvironmentManager) setDefaultResources(env *Environment) {
	switch env.Type {
	case EnvironmentProduction:
		env.Resources = ResourceAllocation{
			MinCPU:      "500m",
			MaxCPU:      "2000m",
			MinMemory:   "512Mi",
			MaxMemory:   "2Gi",
			MinReplicas: 2,
			MaxReplicas: 10,
			StorageSize: "50Gi",
			AutoScale:   true,
		}
	case EnvironmentStaging:
		env.Resources = ResourceAllocation{
			MinCPU:      "250m",
			MaxCPU:      "1000m",
			MinMemory:   "256Mi",
			MaxMemory:   "1Gi",
			MinReplicas: 1,
			MaxReplicas: 3,
			StorageSize: "20Gi",
			AutoScale:   true,
		}
	case EnvironmentDevelopment, EnvironmentPreview:
		env.Resources = ResourceAllocation{
			MinCPU:      "100m",
			MaxCPU:      "500m",
			MinMemory:   "128Mi",
			MaxMemory:   "512Mi",
			MinReplicas: 1,
			MaxReplicas: 1,
			StorageSize: "10Gi",
			AutoScale:   false,
		}
	default:
		env.Resources = ResourceAllocation{
			MinCPU:      "250m",
			MaxCPU:      "1000m",
			MinMemory:   "256Mi",
			MaxMemory:   "1Gi",
			MinReplicas: 1,
			MaxReplicas: 2,
			StorageSize: "20Gi",
			AutoScale:   false,
		}
	}
}
func (em *EnvironmentManager) encryptSecrets(env *Environment) error {
	for key, value := range env.Secrets {
		encrypted, err := em.encrypt(value)
		if err != nil {
			return err
		}
		env.Secrets[key] = encrypted
	}
	return nil
}
func (em *EnvironmentManager) decryptSecrets(env *Environment) error {
	for key, encrypted := range env.Secrets {
		decrypted, err := em.decrypt(encrypted)
		if err != nil {
			return err
		}
		env.Secrets[key] = decrypted
	}
	return nil
}
func (em *EnvironmentManager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(em.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
func (em *EnvironmentManager) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(em.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
func (em *EnvironmentManager) saveEnvironment(env *Environment) error {
	if err := os.MkdirAll(em.storagePath, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	envPath := filepath.Join(em.storagePath, env.ID+".json")
	return os.WriteFile(envPath, data, 0600)
}
func (em *EnvironmentManager) loadEnvironment(envID string) (*Environment, error) {
	envPath := filepath.Join(em.storagePath, envID+".json")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, err
	}
	var env Environment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
func generateID() string {
	return fmt.Sprintf("env_%d", time.Now().UnixNano())
}
