package repositories

import (
	"context"
	"database/sql"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) CreateFollow(ctx context.Context, recipientID, actorID string) error {
	if recipientID == actorID {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (recipient_id, actor_id, type)
		VALUES ($1, $2, 'follow')
		ON CONFLICT DO NOTHING
	`, recipientID, actorID)
	if err != nil {
		return classifyPostgresError(err)
	}
	return nil
}

func (r *NotificationRepository) CreateLike(ctx context.Context, recipientID, actorID, postID string) error {
	if recipientID == actorID {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (recipient_id, actor_id, type, post_id)
		VALUES ($1, $2, 'like', $3)
		ON CONFLICT DO NOTHING
	`, recipientID, actorID, postID)
	if err != nil {
		return classifyPostgresError(err)
	}
	return nil
}

func (r *NotificationRepository) List(ctx context.Context, recipientID string) ([]models.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT notifications.id, notifications.type, users.id, users.handle, users.display_name, notifications.post_id, notifications.created_at
		FROM notifications
		INNER JOIN users ON users.id = notifications.actor_id
		WHERE notifications.recipient_id = $1
		ORDER BY notifications.created_at DESC
	`, recipientID)
	if err != nil {
		return nil, classifyPostgresError(err)
	}
	defer rows.Close()

	notifications := make([]models.Notification, 0)
	for rows.Next() {
		var notification models.Notification
		var postID sql.NullString
		if err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Actor.ID,
			&notification.Actor.Handle,
			&notification.Actor.DisplayName,
			&postID,
			&notification.CreatedAt,
		); err != nil {
			return nil, classifyPostgresError(err)
		}
		if postID.Valid {
			notification.PostID = &postID.String
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, classifyPostgresError(err)
	}
	return notifications, nil
}
