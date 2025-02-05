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
func (s *Storage) SaveComplaint(userUUID string, categoryID int, message string) (string, error) {
	// Сохраняем жалобу
	query := `INSERT INTO complaints (user_uuid, category_id, message) VALUES ($1, $2, $3) RETURNING id`
	var complaintID int
	err := s.db.QueryRow(context.Background(), query, userUUID, categoryID, message).Scan(&complaintID)
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
func (s *Storage) GetComplaintById(complaintID int) (domain.Complaint, error) {
	const op = "storage.postgres.GetComplaintByID"

	query := `
		SELECT id, user_uuid, message, category_id, status, created_at, answer
		FROM complaints
		WHERE id = $1`

	var complaint domain.Complaint
	err := s.db.QueryRow(context.Background(), query, complaintID).Scan(
		&complaint.ID,
		&complaint.UserUUID,
		&complaint.Message,
		&complaint.CategoryID,
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

func (s *Storage) GetComplaints() ([]domain.Complaint, error) {
	const op = "storage.postgres.GetComplaints"

	rows, err := s.db.Query(context.Background(), "SELECT id, user_uuid, message, category_id, status, created_at, answer FROM complaints")
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
			&complaint.CategoryID,
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

func (s *Storage) GetComplaintsByCategoryId(categoryId int) ([]domain.Complaint, error) {
	const op = "storage.postgres.GetComplaintsByCategory"

	rows, err := s.db.Query(context.Background(), "SELECT id, user_uuid, message, category_id, status, created_at FROM complaints WHERE category_id = $1", categoryId)
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
			&complaint.CategoryID,
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
func (s *Storage) UpdateComplaintStatus(complaintID int64, status domain.ComplaintStatus, answer string) error {
	const op = "storage.postgres.UpdateComplaintStatus"

	// Проверяем, существует ли жалоба с таким ID
	var exists bool
	err := s.db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM complaints WHERE id = $1)", complaintID).Scan(&exists)
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
	result, err := s.db.Exec(context.Background(), query, status, complaintID, answer)
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
func (s *Storage) CheckComplaintLimit(userUUID string) (bool, error) {
	const op = "storage.postgres.CheckComplaintLimit"

	// Получаем время последней жалобы от пользователя
	var lastComplaintTime time.Time
	err := s.db.QueryRow(context.Background(), `
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

// DeleteComplaint удаляет жалобу по её ID.
func (s *Storage) DeleteComplaint(id int) error {
	const op = "storage.postgres.DeleteComplaint"

	query := `DELETE FROM complaints WHERE id = $1`
	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
