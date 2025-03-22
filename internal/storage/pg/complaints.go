package pg

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SaveComplaint сохраняет новую жалобу в базе данных.
func (s *Storage) SaveComplaint(ctx context.Context, userUUID string, categoryID int, message string) (string, error) {
	// Сохраняем жалобу
	query := `INSERT INTO complaints (user_uuid, category_id, message) VALUES ($1, $2, $3) RETURNING id`
	var complaintID int
	err := s.db.QueryRow(ctx, query, userUUID, categoryID, message).Scan(&complaintID)
	if err != nil {
		return "", fmt.Errorf("failed to save complaint: %w", err)
	}

	// Получаем ответ из категории
	var answer string
	err = s.db.QueryRow(context.Background(), `SELECT answer FROM categories WHERE id = $1`, categoryID).Scan(&answer)
	if err != nil {
		return "", fmt.Errorf("failed to get category answer: %w", err)
	}

	return answer, nil
}

// GetComplaintById возвращает жалобу по её ID.
func (s *Storage) GetComplaintById(ctx context.Context, complaintID int) (domain.Complaint, error) {
	const op = "storage.postgres.GetComplaintByID"

	query := `
		SELECT c.id, c.user_uuid, c.message, 
		       cat.id, cat.title, cat.description, cat.answer, 
		       c.status, c.created_at, c.answer
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.id
		WHERE c.id = $1`

	var complaint domain.Complaint
	err := s.db.QueryRow(ctx, query, complaintID).Scan(
		&complaint.ID,
		&complaint.UserUUID,
		&complaint.Message,
		&complaint.Category.ID,
		&complaint.Category.Title,
		&complaint.Category.Description,
		&complaint.Category.Answer,
		&complaint.Status,
		&complaint.CreatedAt,
		&complaint.Answer,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Complaint{}, storage.ErrComplaintNotFound
	}
	if err != nil {
		return domain.Complaint{}, fmt.Errorf("%s: %w", op, err)
	}

	return complaint, nil
}

func (s *Storage) GetComplaints(ctx context.Context) ([]domain.Complaint, error) {
	const op = "storage.postgres.GetComplaints"

	query := `
		SELECT c.id, c.user_uuid, c.message, 
		       cat.id, cat.title, cat.description, cat.answer, 
		       c.status, c.created_at, c.answer
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.id`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var complaints []domain.Complaint
	for rows.Next() {
		var complaint domain.Complaint
		err := rows.Scan(
			&complaint.ID,
			&complaint.UserUUID,
			&complaint.Message,
			&complaint.Category.ID,
			&complaint.Category.Title,
			&complaint.Category.Description,
			&complaint.Category.Answer,
			&complaint.Status,
			&complaint.CreatedAt,
			&complaint.Answer,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		complaints = append(complaints, complaint)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return complaints, nil
}

func (s *Storage) GetComplaintsByCategoryId(ctx context.Context, categoryId int) ([]domain.Complaint, error) {
	const op = "storage.postgres.GetComplaintsByCategory"

	rows, err := s.db.Query(ctx, "SELECT id, user_uuid, message, category_id, status, created_at FROM complaints WHERE category_id = $1", categoryId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var complaints []domain.Complaint
	for rows.Next() {
		var complaint domain.Complaint
		err := rows.Scan(
			&complaint.ID,
			&complaint.UserUUID,
			&complaint.Message,
			&complaint.Category,
			&complaint.Status,
			&complaint.CreatedAt,
			&complaint.Answer,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		complaints = append(complaints, complaint)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return complaints, nil
}

// UpdateComplaintStatus обновляет статус жалобы.
func (s *Storage) UpdateComplaintStatus(ctx context.Context, complaintID int64, status domain.ComplaintStatus, answer string) error {
	const op = "storage.postgres.UpdateComplaintStatus"

	// Проверяем, существует ли жалоба с таким ID
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM complaints WHERE id = $1)", complaintID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		// Если жалоба с таким ID не найдена
		return storage.ErrComplaintNotFound
	}

	// Выполняем обновление статуса
	query := `
		UPDATE complaints
		SET status = $1, updated_at = CURRENT_TIMESTAMP, answer = $3
		WHERE id = $2`

	// Выполняем запрос на обновление
	result, err := s.db.Exec(ctx, query, status, complaintID, answer)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем, были ли обновлены строки
	rowsAffected := result.RowsAffected()

	// Если строки не обновлены, значит жалоба с таким ID не существует
	if rowsAffected == 0 {
		return storage.ErrComplaintNotFound
	}

	return nil
}

// CheckComplaintLimit проверяет временной лимит для отправки жалоб.
func (s *Storage) CheckComplaintLimit(ctx context.Context, userUUID string) (bool, error) {
	const op = "storage.postgres.CheckComplaintLimit"

	// Получаем время последней жалобы от пользователя
	var lastComplaintTime time.Time
	err := s.db.QueryRow(ctx, `
		SELECT created_at 
		FROM complaints 
		WHERE user_uuid = $1 
		ORDER BY created_at DESC 
		LIMIT 1`, userUUID).Scan(&lastComplaintTime)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}

	// Проверяем, прошло ли достаточно времени с последней жалобы (например, 1 час)
	if time.Since(lastComplaintTime) < time.Hour {
		return false, storage.ErrLimitOneComplaintInOneHour
	}

	return true, nil
}

// UpdateComplaint обновляет жалобу
func (s *Storage) UpdateComplaint(ctx context.Context, complaintID int64, complaint domain.Complaint) error {
	const op = "storage.postgres.UpdateComplaint"

	// Проверяем, существует ли жалоба с таким ID
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM complaints WHERE id = $1)", complaintID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		// Если жалоба с таким ID не найдена
		return storage.ErrComplaintNotFound
	}

	// Выполняем обновление статуса
	query := `
		UPDATE complaints
		SET id = $1, user_uuid = $2, category_id = $3, message = $4, status = $5, updated_at = CURRENT_TIMESTAMP, answer = $6
		WHERE id = $7`

	// Выполняем запрос на обновление
	result, err := s.db.Exec(ctx, query, complaint.ID, complaint.UserUUID, complaint.Category.ID, complaint.Message, complaint.Status, complaint.Answer.String, complaintID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем, были ли обновлены строки
	rowsAffected := result.RowsAffected()

	// Если строки не обновлены, значит жалоба с таким ID не существует
	if rowsAffected == 0 {
		return storage.ErrComplaintNotFound
	}

	return nil
}

// DeleteComplaint удаляет жалобу по её ID.
func (s *Storage) DeleteComplaint(ctx context.Context, id int) error {
	const op = "storage.postgres.DeleteComplaint"

	query := `DELETE FROM complaints WHERE id = $1`
	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
