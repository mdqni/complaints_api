package pg

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/repository"
	"complaint_server/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type complaintRepo struct {
	db *pgxpool.Pool
}

func NewComplaintRepo(db *Storage) repository.ComplaintRepository {
	return &complaintRepo{db: db.db}
}

func (c complaintRepo) SaveComplaint(ctx context.Context, barcode int, categoryID uuid.UUID, message string) (uuid.UUID, string, error) {
	query := `INSERT INTO complaints (barcode, category_id, message) VALUES ($1, $2, $3) RETURNING uuid`
	var complaintID uuid.UUID
	err := c.db.QueryRow(ctx, query, barcode, categoryID, message).Scan(&complaintID)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("failed to save complaint: %w", err)
	}

	var answer string
	err = c.db.QueryRow(ctx, `SELECT answer FROM categories WHERE uuid = $1`, categoryID).Scan(&answer)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("failed to get categories answer: %w", err)
	}

	return complaintID, answer, nil
}

func (c complaintRepo) IsOwnerOfComplaint(ctx context.Context, id uuid.UUID, barcode int) (bool, error) {
	query := `SELECT barcode FROM complaints WHERE uuid = $1`
	var complaintBarcode int
	err := c.db.QueryRow(ctx, query, id).Scan(&complaintBarcode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrComplaintNotFound
		}
		return false, fmt.Errorf("failed to check complaint ownership: %w", err)
	}
	return complaintBarcode == barcode, nil
}

func (c complaintRepo) GetComplaintByUUID(ctx context.Context, id uuid.UUID) (domain.Complaint, error) {
	const op = "storage.postgres.GetComplaintByUUID"
	query := `
		SELECT c.uuid, c.barcode, c.message, c.status, c.created_at, c.answer,
		       cat.uuid, cat.title, cat.description, cat.answer
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.uuid
		WHERE c.uuid = $1`

	var complaint domain.Complaint
	var category domain.Category

	err := c.db.QueryRow(ctx, query, id).Scan(
		&complaint.ID,
		&complaint.Barcode,
		&complaint.Message,
		&complaint.Status,
		&complaint.CreatedAt,
		&complaint.Answer,
		&category.ID,
		&category.Title,
		&category.Description,
		&category.Answer,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Complaint{}, storage.ErrComplaintNotFound
		}
		fmt.Println("op", op, "err:", err)
		return domain.Complaint{}, storage.ErrScanFailure
	}

	complaint.Category = category
	return complaint, nil
}

func (c complaintRepo) GetComplaints(ctx context.Context) ([]domain.Complaint, error) {
	const op = "storage.postgres.GetComplaints"

	query := `
		SELECT c.uuid, c.barcode, c.message, c.status, c.created_at, c.answer,
		       cat.uuid, cat.title, cat.description, cat.answer
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.uuid`

	rows, err := c.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var complaints []domain.Complaint
	for rows.Next() {
		var complaint domain.Complaint
		var category domain.Category

		err := rows.Scan(
			&complaint.ID,
			&complaint.Barcode,
			&complaint.Message,
			&complaint.Status,
			&complaint.CreatedAt,
			&complaint.Answer,
			&category.ID,
			&category.Title,
			&category.Description,
			&category.Answer,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		complaint.Category = category
		complaints = append(complaints, complaint)
	}

	if len(complaints) == 0 {
		return nil, storage.ErrComplaintNotFound
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return complaints, nil
}

func (c complaintRepo) GetComplaintsByCategoryId(ctx context.Context, categoryId uuid.UUID) ([]domain.Complaint, error) {
	const op = "storage.postgres.GetComplaintsByCategory"

	query := `
		SELECT c.uuid, c.barcode, c.message, c.status, c.created_at, c.answer,
		       cat.uuid, cat.title, cat.description, cat.answer
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.uuid
		WHERE c.category_id = $1`

	rows, err := c.db.Query(ctx, query, categoryId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var complaints []domain.Complaint
	for rows.Next() {
		var complaint domain.Complaint
		var category domain.Category

		err := rows.Scan(
			&complaint.ID,
			&complaint.Barcode,
			&complaint.Message,
			&complaint.Status,
			&complaint.CreatedAt,
			&complaint.Answer,
			&category.ID,
			&category.Title,
			&category.Description,
			&category.Answer,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		complaint.Category = category
		complaints = append(complaints, complaint)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return complaints, nil
}

func (c complaintRepo) GetComplaintsByBarcode(ctx context.Context, barcode int) ([]domain.Complaint, error) {
	rows, err := c.db.Query(ctx, `
	SELECT c.uuid, c.barcode, cat.uuid, cat.title, cat.description, cat.answer,
		c.message, c.status, c.created_at, c.updated_at, c.answer FROM complaints c
	JOIN categories cat ON cat.uuid = c.category_id WHERE c.barcode = $1`, barcode)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	defer rows.Close()

	var complaints []domain.Complaint
	for rows.Next() {
		var c domain.Complaint
		var category domain.Category

		if err := rows.Scan(
			&c.ID,
			&c.Barcode,
			&category.ID,
			&category.Title,
			&category.Description,
			&category.Answer,
			&c.Message,
			&c.Status,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.Answer,
		); err != nil {
			fmt.Println("scan error:", err)
			return nil, err
		}

		c.Category = category
		complaints = append(complaints, c)
	}

	return complaints, nil
}

func (c complaintRepo) CheckComplaintLimit(ctx context.Context, barcode int) (bool, error) {
	const op = "storage.postgres.CheckComplaintLimit"

	var lastComplaintTime time.Time
	err := c.db.QueryRow(ctx, `
		SELECT created_at 
		FROM complaints 
		WHERE barcode = $1 
		ORDER BY created_at DESC 
		LIMIT 1`, barcode).Scan(&lastComplaintTime)

	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}

	if time.Since(lastComplaintTime) < time.Hour {
		return false, storage.ErrLimitOneComplaintInOneHour
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (c complaintRepo) UpdateComplaintStatus(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, answer string) error {
	const op = "storage.postgres.UpdateComplaintStatus"

	var exists bool
	err := c.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM complaints WHERE uuid = $1)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		// Если жалоба с таким ID не найдена
		return storage.ErrComplaintNotFound
	}

	query := `
		UPDATE complaints
		SET status = $1, updated_at = CURRENT_TIMESTAMP, answer = $3
		WHERE uuid = $2`

	result, err := c.db.Exec(ctx, query, status, id, answer)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return storage.ErrComplaintNotFound
	}

	return nil
}

func (c complaintRepo) DeleteComplaint(ctx context.Context, id uuid.UUID) error {
	const op = "storage.postgres.DeleteComplaintById"

	query := `DELETE FROM complaints WHERE uuid = $1`
	r, err := c.db.Exec(ctx, query, id)
	affected := r.RowsAffected()
	if affected == 0 {
		return storage.ErrComplaintNotFound
	}

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c complaintRepo) UpdateComplaint(ctx context.Context, id uuid.UUID, complaint domain.Complaint) (domain.Complaint, error) {
	const op = "storage.postgres.UpdateComplaint"

	var exists bool
	err := c.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM complaints WHERE uuid = $1)", id).Scan(&exists)
	if err != nil {
		return domain.Complaint{}, fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return domain.Complaint{}, storage.ErrComplaintNotFound
	}

	var categoryExists bool
	err = c.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE uuid = $1)", complaint.Category.ID).Scan(&categoryExists)
	if err != nil {
		return domain.Complaint{}, fmt.Errorf("%s: %w", op, err)
	}
	if !categoryExists {
		return domain.Complaint{}, fmt.Errorf("categories with id %s does not exist", complaint.Category.ID.String())
	}

	_, err = c.db.Exec(ctx, `
		UPDATE complaints
		SET barcode = $1, category_id = $2, message = $3, status = $4, answer = $5, updated_at = $6
		WHERE uuid = $7
	`,
		complaint.Barcode,
		complaint.Category.ID,
		complaint.Message,
		complaint.Status,
		complaint.Answer.String,
		time.Now(),
		id,
	)
	if err != nil {
		return domain.Complaint{}, fmt.Errorf("%s: %w", op, err)
	}

	query := `
		SELECT 
			c.uuid, c.barcode, c.category_id, c.message, c.status, c.answer, c.created_at, c.updated_at,
			cat.title, cat.description, cat.answer
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.uuid
		WHERE c.uuid = $1
	`

	var updated domain.Complaint
	var answer sql.NullString

	err = c.db.QueryRow(ctx, query, id).Scan(
		&updated.ID,
		&updated.Barcode,
		&updated.Category.ID,
		&updated.Message,
		&updated.Status,
		&answer,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.Category.Title,
		&updated.Category.Description,
		&updated.Category.Answer,
	)
	if err != nil {
		return domain.Complaint{}, fmt.Errorf("%s: %w", op, err)
	}

	updated.Answer = answer
	return updated, nil
}
