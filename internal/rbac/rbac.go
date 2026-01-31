package rbac

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidRole  = errors.New("invalid role")
)

// Role represents a user role
type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleViewer    Role = "viewer"
)

// Permission represents an action permission
type Permission string

const (
	PermProjectCreate  Permission = "project:create"
	PermProjectRead    Permission = "project:read"
	PermProjectUpdate  Permission = "project:update"
	PermProjectDelete  Permission = "project:delete"
	PermDeployCreate   Permission = "deploy:create"
	PermDeployRollback Permission = "deploy:rollback"
	PermSettingsUpdate Permission = "settings:update"
	PermBillingView    Permission = "billing:view"
	PermBillingUpdate  Permission = "billing:update"
	PermMemberInvite   Permission = "member:invite"
	PermMemberRemove   Permission = "member:remove"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleOwner: {
		PermProjectCreate, PermProjectRead, PermProjectUpdate, PermProjectDelete,
		PermDeployCreate, PermDeployRollback,
		PermSettingsUpdate, PermBillingView, PermBillingUpdate,
		PermMemberInvite, PermMemberRemove,
	},
	RoleAdmin: {
		PermProjectCreate, PermProjectRead, PermProjectUpdate, PermProjectDelete,
		PermDeployCreate, PermDeployRollback,
		PermSettingsUpdate, PermMemberInvite,
	},
	RoleDeveloper: {
		PermProjectRead, PermProjectUpdate,
		PermDeployCreate, PermDeployRollback,
	},
	RoleViewer: {
		PermProjectRead,
	},
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	UserID         string                 `json:"user_id"`
	UserEmail      string                 `json:"user_email"`
	Action         string                 `json:"action"`
	ResourceType   string                 `json:"resource_type"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// RBACService handles role-based access control
type RBACService struct {
	db *sql.DB
}

// NewRBACService creates a new RBAC service
func NewRBACService(db *sql.DB) *RBACService {
	return &RBACService{db: db}
}

// CheckPermission verifies if a user has a specific permission
func (rs *RBACService) CheckPermission(ctx context.Context, userID, orgID string, permission Permission) error {
	// Get user's role in organization
	var role string
	err := rs.db.QueryRowContext(ctx, `
		SELECT role FROM organization_members
		WHERE user_id = $1 AND organization_id = $2
	`, userID, orgID).Scan(&role)

	if err != nil {
		return ErrUnauthorized
	}

	// Check if role has permission
	permissions, ok := RolePermissions[Role(role)]
	if !ok {
		return ErrInvalidRole
	}

	for _, p := range permissions {
		if p == permission {
			return nil
		}
	}

	return ErrUnauthorized
}

// GetUserRole returns a user's role in an organization
func (rs *RBACService) GetUserRole(ctx context.Context, userID, orgID string) (Role, error) {
	var role string
	err := rs.db.QueryRowContext(ctx, `
		SELECT role FROM organization_members
		WHERE user_id = $1 AND organization_id = $2
	`, userID, orgID).Scan(&role)

	if err != nil {
		return "", err
	}

	return Role(role), nil
}

// UpdateUserRole updates a user's role in an organization
func (rs *RBACService) UpdateUserRole(ctx context.Context, targetUserID, orgID string, newRole Role) error {
	_, err := rs.db.ExecContext(ctx, `
		UPDATE organization_members
		SET role = $1
		WHERE user_id = $2 AND organization_id = $3
	`, newRole, targetUserID, orgID)

	return err
}

// LogAction creates an audit log entry
func (rs *RBACService) LogAction(ctx context.Context, log *AuditLog) error {
	metadataJSON, _ := json.Marshal(log.Metadata)

	return rs.db.QueryRowContext(ctx, `
		INSERT INTO audit_logs (organization_id, user_id, user_email, action, resource_type, resource_id, ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`, log.OrganizationID, log.UserID, log.UserEmail, log.Action, log.ResourceType,
		log.ResourceID, log.IPAddress, log.UserAgent, metadataJSON).
		Scan(&log.ID, &log.CreatedAt)
}

// GetAuditLogs retrieves audit logs for an organization
func (rs *RBACService) GetAuditLogs(ctx context.Context, orgID string, limit int) ([]AuditLog, error) {
	rows, err := rs.db.QueryContext(ctx, `
		SELECT id, organization_id, user_id, user_email, action, resource_type, resource_id, 
		       ip_address, user_agent, metadata, created_at
		FROM audit_logs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var resourceID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(&log.ID, &log.OrganizationID, &log.UserID, &log.UserEmail,
			&log.Action, &log.ResourceType, &resourceID, &log.IPAddress, &log.UserAgent,
			&metadataJSON, &log.CreatedAt)
		if err != nil {
			continue
		}

		if resourceID.Valid {
			log.ResourceID = resourceID.String
		}
		json.Unmarshal(metadataJSON, &log.Metadata)

		logs = append(logs, log)
	}

	return logs, nil
}

// InviteMember invites a new member to an organization
func (rs *RBACService) InviteMember(ctx context.Context, orgID, email string, role Role) error {
	// Create invitation
	_, err := rs.db.ExecContext(ctx, `
		INSERT INTO organization_invitations (organization_id, email, role, expires_at)
		VALUES ($1, $2, $3, NOW() + INTERVAL '7 days')
	`, orgID, email, role)

	// TODO: Send invitation email

	return err
}

// AcceptInvitation accepts an organization invitation
func (rs *RBACService) AcceptInvitation(ctx context.Context, invitationID, userID string) error {
	// Get invitation details
	var orgID string
	var role string
	var expiresAt time.Time

	err := rs.db.QueryRowContext(ctx, `
		SELECT organization_id, role, expires_at
		FROM organization_invitations
		WHERE id = $1 AND status = 'pending'
	`, invitationID).Scan(&orgID, &role, &expiresAt)

	if err != nil {
		return errors.New("invalid invitation")
	}

	if time.Now().After(expiresAt) {
		return errors.New("invitation expired")
	}

	// Add user to organization
	_, err = rs.db.ExecContext(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
	`, orgID, userID, role)

	if err != nil {
		return err
	}

	// Mark invitation as accepted
	_, err = rs.db.ExecContext(ctx, `
		UPDATE organization_invitations
		SET status = 'accepted', accepted_at = NOW()
		WHERE id = $1
	`, invitationID)

	return err
}

// RemoveMember removes a member from an organization
func (rs *RBACService) RemoveMember(ctx context.Context, orgID, userID string) error {
	_, err := rs.db.ExecContext(ctx, `
		DELETE FROM organization_members
		WHERE organization_id = $1 AND user_id = $2 AND role != 'owner'
	`, orgID, userID)

	return err
}
