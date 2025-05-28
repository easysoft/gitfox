// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types"
	"github.com/easysoft/gitfox/types/enum"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ store.ConnectorStore = (*connectorStore)(nil)

const (
	//nolint:goconst
	connectorQueryBase = `
		SELECT` + connectorColumns + `
		FROM connectors`

	connectorColumns = `
	connector_id,
	connector_identifier,
	connector_description,
	connector_type,
	connector_auth_type,
	connector_created_by,
	connector_space_id,
	connector_last_test_attempt,
	connector_last_test_error_msg,
	connector_last_test_status,
	connector_created,
	connector_updated,
	connector_version,
	connector_address,
	connector_insecure,
	connector_username,
	connector_github_app_installation_id,
	connector_github_app_application_id,
	connector_region,
	connector_password,
	connector_token,
	connector_aws_key,
	connector_aws_secret,
	connector_github_app_private_key,
	connector_token_refresh
	`
)

type connector struct {
	ID               int64  `gorm:"column:connector_id"`
	Identifier       string `gorm:"column:connector_identifier"`
	Description      string `gorm:"column:connector_description"`
	Type             string `gorm:"column:connector_type"`
	AuthType         string `gorm:"column:connector_auth_type"`
	CreatedBy        int64  `gorm:"column:connector_created_by"`
	SpaceID          int64  `gorm:"column:connector_space_id"`
	LastTestAttempt  int64  `gorm:"column:connector_last_test_attempt"`
	LastTestErrorMsg string `gorm:"column:connector_last_test_error_msg"`
	LastTestStatus   string `gorm:"column:connector_last_test_status"`
	Created          int64  `gorm:"column:connector_created"`
	Updated          int64  `gorm:"column:connector_updated"`
	Version          int64  `gorm:"column:connector_version"`

	Address                 sql.NullString `gorm:"column:connector_address"`
	Insecure                sql.NullBool   `gorm:"column:connector_insecure"`
	Username                sql.NullString `gorm:"column:connector_username"`
	GithubAppInstallationID sql.NullString `gorm:"column:connector_github_app_installation_id"`
	GithubAppApplicationID  sql.NullString `gorm:"column:connector_github_app_application_id"`
	Region                  sql.NullString `gorm:"column:connector_region"`
	// Password fields are stored as reference to secrets table
	Password            sql.NullInt64 `gorm:"column:connector_password"`
	Token               sql.NullInt64 `gorm:"column:connector_token"`
	AWSKey              sql.NullInt64 `gorm:"column:connector_aws_key"`
	AWSSecret           sql.NullInt64 `gorm:"column:connector_aws_secret"`
	GithubAppPrivateKey sql.NullInt64 `gorm:"column:connector_github_app_private_key"`
	TokenRefresh        sql.NullInt64 `gorm:"column:connector_token_refresh"`
}

// NewConnectorStore returns a new ConnectorStore.
// The secret store is used to resolve the secret references.
func NewConnectorStore(db *gorm.DB, secretStore store.SecretStore) store.ConnectorStore {
	return &connectorStore{
		db:          db,
		secretStore: secretStore,
	}
}

func (s *connectorStore) mapFromDBConnectors(ctx context.Context, src []*connector) ([]*types.Connector, error) {
	dst := make([]*types.Connector, len(src))
	for i, v := range src {
		m, err := s.mapFromDBConnector(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("could not map from db connector: %w", err)
		}
		dst[i] = m
	}
	return dst, nil
}

func (s *connectorStore) mapToDBConnector(ctx context.Context, v *types.Connector) (*connector, error) {
	to := connector{
		ID:               v.ID,
		Identifier:       v.Identifier,
		Description:      v.Description,
		Type:             v.Type.String(),
		SpaceID:          v.SpaceID,
		CreatedBy:        v.CreatedBy,
		Created:          v.Created,
		Updated:          v.Updated,
		Version:          v.Version,
		LastTestAttempt:  v.LastTestAttempt,
		LastTestErrorMsg: v.LastTestErrorMsg,
		LastTestStatus:   v.LastTestStatus.String(),
	}
	// Parse connector specific configs
	err := s.convertConfigToDB(ctx, v, &to)
	if err != nil {
		return nil, fmt.Errorf("could not convert config to db: %w", err)
	}
	return &to, nil
}

func (s *connectorStore) convertConfigToDB(
	ctx context.Context,
	source *types.Connector,
	to *connector,
) error {
	switch {
	case source.Github != nil:
		to.Address = sql.NullString{String: source.Github.APIURL, Valid: true}
		to.Insecure = sql.NullBool{Bool: source.Github.Insecure, Valid: true}
		if source.Github.Auth == nil {
			return fmt.Errorf("auth is required for github connectors")
		}
		if source.Github.Auth.AuthType != enum.ConnectorAuthTypeBearer {
			return fmt.Errorf("only bearer token auth is supported for github connectors")
		}
		to.AuthType = source.Github.Auth.AuthType.String()
		creds := source.Github.Auth.Bearer
		// use the same space ID as the connector
		tokenID, err := s.secretIdentiferToID(ctx, creds.Token.Identifier, source.SpaceID)
		if err != nil {
			return fmt.Errorf("could not find secret: %w", err)
		}
		to.Token = sql.NullInt64{Int64: tokenID, Valid: true}
	default:
		return fmt.Errorf("no connector config found for type: %s", source.Type)
	}
	return nil
}

// secretIdentiferToID finds the secret ID given the space ID and the identifier.
func (s *connectorStore) secretIdentiferToID(
	ctx context.Context,
	identifier string,
	spaceID int64,
) (int64, error) {
	secret, err := s.secretStore.FindByIdentifier(ctx, spaceID, identifier)
	if err != nil {
		return 0, err
	}
	return secret.ID, nil
}

func (s *connectorStore) mapFromDBConnector(
	ctx context.Context,
	dbConnector *connector,
) (*types.Connector, error) {
	connector := &types.Connector{
		ID:               dbConnector.ID,
		Identifier:       dbConnector.Identifier,
		Description:      dbConnector.Description,
		Type:             enum.ConnectorType(dbConnector.Type),
		SpaceID:          dbConnector.SpaceID,
		CreatedBy:        dbConnector.CreatedBy,
		LastTestAttempt:  dbConnector.LastTestAttempt,
		LastTestErrorMsg: dbConnector.LastTestErrorMsg,
		LastTestStatus:   enum.ConnectorStatus(dbConnector.LastTestStatus),
		Created:          dbConnector.Created,
		Updated:          dbConnector.Updated,
		Version:          dbConnector.Version,
	}
	err := s.populateConnectorData(ctx, dbConnector, connector)
	if err != nil {
		return nil, fmt.Errorf("could not populate connector data: %w", err)
	}
	return connector, nil
}

func (s *connectorStore) populateConnectorData(
	ctx context.Context,
	source *connector,
	to *types.Connector,
) error {
	switch enum.ConnectorType(source.Type) {
	case enum.ConnectorTypeGithub:
		githubData, err := s.parseGithubConnectorData(ctx, source)
		if err != nil {
			return fmt.Errorf("could not parse github connector data: %w", err)
		}
		to.Github = githubData
	// Cases for other connectors can be added here
	default:
		return fmt.Errorf("unsupported connector type: %s", source.Type)
	}
	return nil
}

func (s *connectorStore) parseGithubConnectorData(
	ctx context.Context,
	connector *connector,
) (*types.GithubConnectorData, error) {
	auth, err := s.parseAuthenticationData(ctx, connector)
	if err != nil {
		return nil, fmt.Errorf("could not parse authentication data: %w", err)
	}
	return &types.GithubConnectorData{
		APIURL:   connector.Address.String,
		Insecure: connector.Insecure.Bool,
		Auth:     auth,
	}, nil
}

func (s *connectorStore) parseAuthenticationData(
	ctx context.Context,
	connector *connector,
) (*types.ConnectorAuth, error) {
	authType, err := enum.ParseConnectorAuthType(connector.AuthType)
	if err != nil {
		return nil, err
	}

	switch authType {
	case enum.ConnectorAuthTypeBasic:
		if !connector.Username.Valid || !connector.Password.Valid {
			return nil, fmt.Errorf("basic auth requires both username and password")
		}
		passwordRef, err := s.convertToRef(ctx, connector.Password.Int64)
		if err != nil {
			return nil, fmt.Errorf("could not convert basicauth password to ref: %w", err)
		}
		return &types.ConnectorAuth{
			AuthType: enum.ConnectorAuthTypeBasic,
			Basic: &types.BasicAuthCreds{
				Username: connector.Username.String,
				Password: passwordRef,
			},
		}, nil
	case enum.ConnectorAuthTypeBearer:
		if !connector.Token.Valid {
			return nil, fmt.Errorf("bearer auth requires a token")
		}
		tokenRef, err := s.convertToRef(ctx, connector.Token.Int64)
		if err != nil {
			return nil, fmt.Errorf("could not convert bearer token to ref: %w", err)
		}
		return &types.ConnectorAuth{
			AuthType: enum.ConnectorAuthTypeBearer,
			Bearer: &types.BearerTokenCreds{
				Token: tokenRef,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", connector.AuthType)
	}
}

func (s *connectorStore) convertToRef(ctx context.Context, id int64) (types.SecretRef, error) {
	secret, err := s.secretStore.Find(ctx, id)
	if err != nil {
		return types.SecretRef{}, err
	}
	return types.SecretRef{
		Identifier: secret.Identifier,
	}, nil
}

type connectorStore struct {
	db          *gorm.DB
	secretStore store.SecretStore
}

// Find returns a connector given a connector ID.
func (s *connectorStore) Find(ctx context.Context, id int64) (*types.Connector, error) {
	dst := new(connector)
	if err := s.db.Where("connector_id = ?", id).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find connector")
	}
	return s.mapFromDBConnector(ctx, dst)
}

// FindByIdentifier returns a connector in a given space with a given identifier.
func (s *connectorStore) FindByIdentifier(
	ctx context.Context,
	spaceID int64,
	identifier string,
) (*types.Connector, error) {
	dst := new(connector)
	if err := s.db.Where("connector_space_id = ? AND connector_identifier = ?", spaceID, identifier).First(dst).Error; err != nil {
		return nil, database.ProcessGormSQLErrorf(ctx, err, "Failed to find connector")
	}
	return s.mapFromDBConnector(ctx, dst)
}

// Create creates a connector.
func (s *connectorStore) Create(ctx context.Context, connector *types.Connector) error {
	dbConnector, err := s.mapToDBConnector(ctx, connector)
	if err != nil {
		return err
	}

	if err := s.db.Create(dbConnector).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "connector query failed")
	}
	connector.ID = dbConnector.ID
	return nil
}

func (s *connectorStore) Update(ctx context.Context, p *types.Connector) error {
	conn, err := s.mapToDBConnector(ctx, p)
	if err != nil {
		return err
	}
	o := *conn
	o.Version++
	o.Updated = time.Now().UnixMilli()

	result := s.db.Model(&connector{}).
		Where("connector_id = ? AND connector_version = ?", o.ID, o.Version-1).
		Updates(map[string]interface{}{
			"connector_description":                o.Description,
			"connector_identifier":                 o.Identifier,
			"connector_last_test_attempt":          o.LastTestAttempt,
			"connector_last_test_error_msg":        o.LastTestErrorMsg,
			"connector_last_test_status":           o.LastTestStatus,
			"connector_updated":                    o.Updated,
			"connector_version":                    o.Version,
			"connector_auth_type":                  o.AuthType,
			"connector_address":                    o.Address,
			"connector_insecure":                   o.Insecure,
			"connector_username":                   o.Username,
			"connector_github_app_installation_id": o.GithubAppInstallationID,
			"connector_github_app_application_id":  o.GithubAppApplicationID,
			"connector_region":                     o.Region,
			"connector_password":                   o.Password,
			"connector_token":                      o.Token,
			"connector_aws_key":                    o.AWSKey,
			"connector_aws_secret":                 o.AWSSecret,
			"connector_github_app_private_key":     o.GithubAppPrivateKey,
			"connector_token_refresh":              o.TokenRefresh,
		})
	if result.Error != nil {
		return database.ProcessGormSQLErrorf(ctx, result.Error, "Failed to update connector")
	}

	count := result.RowsAffected
	if count == 0 {
		return gitfox_store.ErrVersionConflict
	}

	p.Version = o.Version
	p.Updated = o.Updated
	return nil
}

// UpdateOptLock updates the connector using the optimistic locking mechanism.
func (s *connectorStore) UpdateOptLock(ctx context.Context,
	connector *types.Connector,
	mutateFn func(connector *types.Connector) error,
) (*types.Connector, error) {
	for {
		dup := *connector

		err := mutateFn(&dup)
		if err != nil {
			return nil, err
		}

		err = s.Update(ctx, &dup)
		if err == nil {
			return &dup, nil
		}
		if !errors.Is(err, gitfox_store.ErrVersionConflict) {
			return nil, err
		}

		connector, err = s.Find(ctx, connector.ID)
		if err != nil {
			return nil, err
		}
	}
}

// List lists all the connectors present in a space.
func (s *connectorStore) List(
	ctx context.Context,
	parentID int64,
	filter types.ListQueryFilter,
) ([]*types.Connector, error) {
	var connectors []*connector
	err := s.db.
		Table("connectors").
		Select(connectorColumns).
		Where("connector_space_id = ?", parentID).
		Where("LOWER(connector_identifier) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(filter.Query))).
		Limit(int(filter.Size)).
		Offset(int(filter.Page * filter.Size)).
		Find(&connectors).Error
	if err != nil {
		return nil, errors.Wrap(err, "Failed executing custom list query")
	}

	return s.mapFromDBConnectors(ctx, connectors)
}

// Delete deletes a connector given a connector ID.
func (s *connectorStore) Delete(ctx context.Context, id int64) error {
	if err := s.db.Delete(&connector{}, id).Error; err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete connector")
	}

	return nil
}

// DeleteByIdentifier deletes a connector with a given identifier in a space.
func (s *connectorStore) DeleteByIdentifier(ctx context.Context, spaceID int64, identifier string) error {
	err := s.db.
		Table("connectors").
		Where("connector_space_id = ? AND connector_identifier = ?", spaceID, identifier).
		Delete(&connector{}).
		Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Could not delete connector")
	}

	return nil
}

// Count of connectors in a space.
func (s *connectorStore) Count(ctx context.Context, parentID int64, filter types.ListQueryFilter) (int64, error) {
	var count int64
	err := s.db.
		Model(&connector{}).
		Where("connector_space_id = ?", parentID).
		Where("LOWER(connector_identifier) LIKE ?", fmt.Sprintf("%%%s%%", filter.Query)).
		Count(&count).Error
	if err != nil {
		return 0, database.ProcessGormSQLErrorf(ctx, err, "Failed executing count query")
	}
	return count, nil
}
