package accesskeys

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreAccessKey struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreAccessKey {
	return &StoreAccessKey{db}
}

type RevokeCodeData struct {
	RevokeBy uuid.UUID
	code     string
}

func (s *StoreAccessKey) Create(ctx context.Context, AccessKey *model.AccessKey) error {
	query := `
		INSERT INTO access_keys (code, created_by, expires_at)
		VALUES ($1, $2, $3) RETURNING code, id
	`
	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		AccessKey.Code,
		AccessKey.CreatedBy,
		AccessKey.ExpiresAt,
	).Scan(
		&AccessKey.Code,
		&AccessKey.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *StoreAccessKey) GetByID(ctx context.Context, accessKeyID uuid.UUID) (model.AccessKey, error) {
	query := `
		SELECT id, code, created_by, redeemed_by,
		       redeemed_at, expires_at, revoked_at, revoked_by, created_at
		FROM access_keys
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	AccessKey := model.AccessKey{}

	err := s.db.QueryRowContext(ctx, query, accessKeyID).Scan(
		&AccessKey.ID,
		&AccessKey.Code,
		&AccessKey.CreatedBy,
		&AccessKey.RedeemedBy,
		&AccessKey.RedeemedAt,
		&AccessKey.ExpiresAt,
		&AccessKey.RevokedAt,
		&AccessKey.RevokedBy,
		&AccessKey.CreatedAt,
	)

	if err != nil {
		return model.AccessKey{}, err
	}

	return AccessKey, nil
}

func (s *StoreAccessKey) Redeem(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error) {
	query := `
		UPDATE access_keys
		SET redeemed_by = $1, redeemed_at = now()
		WHERE code = $2 AND redeemed_by IS NULL 
			  AND expires_at > now() AND revoked_at IS NULL
		RETURNING id, expires_at
	`

	exec := dbtx.FromContext(ctx, s.db)

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	row := exec.QueryRowContext(ctx, query, userID, code)

	var ak model.AccessKey

	if err := row.Scan(&ak.ID, &ak.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotRedeemable
		}
		return nil, fmt.Errorf("canjeando access key: %w", err)
	}

	return &ak, nil
}

func (s *StoreAccessKey) Revoke(ctx context.Context, revokeData RevokeCodeData) error {
	query := `
		UPDATE access_keys
		SET revoked_by = $1, revoked_at = now()
		WHERE code = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, revokeData.RevokeBy, revokeData.code)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return util.ErrNotFound
	}

	return nil
}
