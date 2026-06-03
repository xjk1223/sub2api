package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	entcheckin "github.com/Wei-Shaw/sub2api/ent/checkinlog"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type checkinRepository struct {
	client *ent.Client
}

// NewCheckinRepository creates the check-in log repository.
func NewCheckinRepository(client *ent.Client) service.CheckinRepository {
	return &checkinRepository{client: client}
}

func (r *checkinRepository) GetByDate(ctx context.Context, userID int64, date time.Time) (*service.CheckinLog, error) {
	m, err := r.client.CheckinLog.Query().
		Where(entcheckin.UserID(userID), entcheckin.CheckinDate(date)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toServiceCheckinLog(m), nil
}

func (r *checkinRepository) GetLastBefore(ctx context.Context, userID int64, before time.Time) (*service.CheckinLog, error) {
	m, err := r.client.CheckinLog.Query().
		Where(entcheckin.UserID(userID), entcheckin.CheckinDateLT(before)).
		Order(ent.Desc(entcheckin.FieldCheckinDate)).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toServiceCheckinLog(m), nil
}

func (r *checkinRepository) Create(ctx context.Context, log *service.CheckinLog) (*service.CheckinLog, error) {
	m, err := r.client.CheckinLog.Create().
		SetUserID(log.UserID).
		SetAmount(log.Amount).
		SetConsecutiveDays(log.ConsecutiveDays).
		SetCheckinDate(log.CheckinDate).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toServiceCheckinLog(m), nil
}

func (r *checkinRepository) List(ctx context.Context, userID int64, page, pageSize int) ([]*service.CheckinLog, int64, error) {
	q := r.client.CheckinLog.Query()
	if userID > 0 {
		q = q.Where(entcheckin.UserID(userID))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.
		Order(ent.Desc(entcheckin.FieldCheckinDate)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	logs := make([]*service.CheckinLog, 0, len(rows))
	for _, m := range rows {
		logs = append(logs, toServiceCheckinLog(m))
	}
	return logs, int64(total), nil
}

func toServiceCheckinLog(m *ent.CheckinLog) *service.CheckinLog {
	return &service.CheckinLog{
		ID:              m.ID,
		UserID:          m.UserID,
		Amount:          m.Amount,
		ConsecutiveDays: m.ConsecutiveDays,
		CheckinDate:     m.CheckinDate,
		CreatedAt:       m.CreatedAt,
	}
}
