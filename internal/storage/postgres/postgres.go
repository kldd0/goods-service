package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kldd0/goods-service/internal/domain/models"
	"github.com/kldd0/goods-service/internal/storage"

	"github.com/jmoiron/sqlx"
)

const dbDriver = "pgx"

type Storage struct {
	db *sqlx.DB
}

func New(dbUri string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Open(dbDriver, dbUri)
	if err != nil {
		return nil, fmt.Errorf("%s: open db connection: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: db ping failed: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) GetGood(ctx context.Context, goodId string, projectId string) (models.Good, error) {
	const op = "storage.postgres.GetGood"

	q := `SELECT * FROM goods WHERE id=$1 AND project_id=$2;`

	stmt, err := s.db.PrepareContext(ctx, q)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resultGood models.Good

	err = stmt.QueryRowContext(ctx, goodId, projectId).Scan(
		&resultGood.ID,
		&resultGood.ProjectId,
		&resultGood.Name,
		&resultGood.Description,
		&resultGood.Priority,
		&resultGood.Removed,
		&resultGood.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Good{}, storage.ErrEntryDoesntExist
		}

		return models.Good{}, fmt.Errorf("%s: saving entry: %w", op, err)
	}

	return resultGood, nil
}

func (s *Storage) SaveGood(ctx context.Context, good models.Good) (models.Good, error) {
	const op = "storage.postgres.SaveGood"

	// check if entry exists
	q := `SELECT id FROM goods WHERE id=$1 FOR SHARE;`

	stmt, err := s.db.PrepareContext(ctx, q)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, good.ID)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		if c != 0 {
			return models.Good{}, storage.ErrEntryAlreadyExists
		}

		return models.Good{}, fmt.Errorf("%s: get info about affected rows: %w", op, err)
	}

	q = `INSERT INTO goods (project_id, name, description, priority, removed, created_at)
            VALUES ($1, $2, $3, $4, $5, $6) RETURNING
            id, project_id, name, description, priority, removed, created_at;`

	stmt, err = s.db.PrepareContext(ctx, q)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resultGood models.Good

	err = stmt.QueryRowContext(
		ctx, good.ProjectId, good.Name, good.Description, good.Priority, good.Removed, time.Now(),
	).Scan(
		&resultGood.ID,
		&resultGood.ProjectId,
		&resultGood.Name,
		&resultGood.Description,
		&resultGood.Priority,
		&resultGood.Removed,
		&resultGood.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Good{}, storage.ErrGettingInsertedRows
		}

		return models.Good{}, fmt.Errorf("%s: saving entry: %w", op, err)
	}

	return resultGood, nil
}

func (s *Storage) PatchGood(ctx context.Context, patchedGood models.Good) (models.Good, error) {
	const op = "storage.postgres.Patch"

	// Begin transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: begin transaction: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
	}()

	// check if entry exists
	q := `SELECT id FROM goods WHERE id=$1 AND project_id=$2 FOR SHARE;`

	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, patchedGood.ID, patchedGood.ProjectId)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Good{}, storage.ErrEntryDoesntExist
		}

		return models.Good{}, fmt.Errorf("%s: get info about affected rows: %w", op, err)
	}

	if c == 0 {
		return models.Good{}, storage.ErrEntryDoesntExist
	}

	q = `UPDATE goods SET name=$1, description=$2, priority=$3, removed=$4 WHERE id=$5 AND project_id=$6 RETURNING
            id, project_id, name, description, priority, removed, created_at;`

	stmt, err = tx.PrepareContext(ctx, q)
	if err != nil {
		return models.Good{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resultGood models.Good

	err = stmt.QueryRowContext(
		ctx,
		patchedGood.Name,
		patchedGood.Description,
		patchedGood.Priority,
		patchedGood.Removed,
		patchedGood.ID,
		patchedGood.ProjectId,
	).Scan(
		&resultGood.ID,
		&resultGood.ProjectId,
		&resultGood.Name,
		&resultGood.Description,
		&resultGood.Priority,
		&resultGood.Removed,
		&resultGood.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Good{}, storage.ErrGettingInsertedRows
		}

		return models.Good{}, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	// commit the transaction
	if err := tx.Commit(); err != nil {
		return models.Good{}, fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return resultGood, nil
}

func (s *Storage) ListGoodsWithPagination(ctx context.Context, offset, limit string) ([]models.Good, error) {
	const op = "storage.postgres.ListGoodsWithPagination"

	q := `SELECT id, project_id, name, description, priority, removed, created_at FROM goods
            ORDER BY id OFFSET $1 LIMIT $2;`

	stmt, err := s.db.PrepareContext(ctx, q)
	if err != nil {
		return []models.Good{}, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, offset, limit)
	if err != nil {
		return []models.Good{}, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	var goods []models.Good

	for rows.Next() {
		var good models.Good

		if err := rows.Scan(
			&good.ID,
			&good.ProjectId,
			&good.Name,
			&good.Description,
			&good.Priority,
			&good.Removed,
			&good.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scanning bytes of row: %w", op, err)
		}

		goods = append(goods, good)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error iterating over rows: %w", op, err)
	}

	return goods, nil
}

func (s *Storage) DeleteGood(ctx context.Context, goodId string, projectId string) error {
	const op = "storage.postgres.DeleteGood"

	q := `DELETE FROM goods WHERE id=$1 AND project_id=$2`

	stmt, err := s.db.PrepareContext(ctx, q)
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, goodId, projectId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrEntryDoesntExist
		}

		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
