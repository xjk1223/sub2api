package service

import (
	"context"
	"errors"
	"strconv"
	"time"
)

var (
	ErrCheckinAlreadyDone = errors.New("already checked in today")
	ErrCheckinDisabled    = errors.New("checkin feature is disabled")
)

// CheckinLog is the domain model for a daily check-in record.
type CheckinLog struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Amount          float64   `json:"amount"`
	ConsecutiveDays int       `json:"consecutive_days"`
	CheckinDate     time.Time `json:"checkin_date"`
	CreatedAt       time.Time `json:"created_at"`
}

// CheckinSettings is persisted in the settings table.
type CheckinSettings struct {
	Enabled          bool    `json:"enabled"`
	BaseAmount       float64 `json:"base_amount"`
	ConsecutiveBonus bool    `json:"consecutive_bonus"`
	BonusPerDay      float64 `json:"bonus_per_day"`
	MaxBonusDays     int     `json:"max_bonus_days"`
}

// CheckinStatus is the per-user check-in status returned to the UI.
type CheckinStatus struct {
	Enabled         bool       `json:"enabled"`
	CheckedInToday  bool       `json:"checked_in_today"`
	ConsecutiveDays int        `json:"consecutive_days"`
	TodayAmount     float64    `json:"today_amount"`
	LastCheckinDate *time.Time `json:"last_checkin_date,omitempty"`
}

// CheckinResult is returned after a successful check-in.
type CheckinResult struct {
	Amount          float64 `json:"amount"`
	ConsecutiveDays int     `json:"consecutive_days"`
	NewBalance      float64 `json:"new_balance"`
}

// CheckinRepository is the data-access interface for check-in logs.
type CheckinRepository interface {
	GetByDate(ctx context.Context, userID int64, date time.Time) (*CheckinLog, error)
	GetLastBefore(ctx context.Context, userID int64, before time.Time) (*CheckinLog, error)
	Create(ctx context.Context, log *CheckinLog) (*CheckinLog, error)
	List(ctx context.Context, userID int64, page, pageSize int) ([]*CheckinLog, int64, error)
}

// CheckinService implements the daily check-in business logic.
type CheckinService struct {
	checkinRepo CheckinRepository
	settingRepo SettingRepository
	adminSvc    AdminService
}

func NewCheckinService(checkinRepo CheckinRepository, settingRepo SettingRepository, adminSvc AdminService) *CheckinService {
	return &CheckinService{checkinRepo: checkinRepo, settingRepo: settingRepo, adminSvc: adminSvc}
}

func (s *CheckinService) GetSettings(ctx context.Context) (*CheckinSettings, error) {
	keys := []string{
		"checkin_enabled", "checkin_base_amount", "checkin_consecutive_bonus",
		"checkin_bonus_per_day", "checkin_max_bonus_days",
	}
	m, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, err
	}
	return &CheckinSettings{
		Enabled:          checkinParseBool(m["checkin_enabled"], true),
		BaseAmount:       checkinParseFloat(m["checkin_base_amount"], 0.1),
		ConsecutiveBonus: checkinParseBool(m["checkin_consecutive_bonus"], true),
		BonusPerDay:      checkinParseFloat(m["checkin_bonus_per_day"], 0.05),
		MaxBonusDays:     checkinParseInt(m["checkin_max_bonus_days"], 7),
	}, nil
}

func (s *CheckinService) UpdateSettings(ctx context.Context, settings *CheckinSettings) error {
	return s.settingRepo.SetMultiple(ctx, map[string]string{
		"checkin_enabled":           strconv.FormatBool(settings.Enabled),
		"checkin_base_amount":       strconv.FormatFloat(settings.BaseAmount, 'f', -1, 64),
		"checkin_consecutive_bonus": strconv.FormatBool(settings.ConsecutiveBonus),
		"checkin_bonus_per_day":     strconv.FormatFloat(settings.BonusPerDay, 'f', -1, 64),
		"checkin_max_bonus_days":    strconv.Itoa(settings.MaxBonusDays),
	})
}

func (s *CheckinService) GetStatus(ctx context.Context, userID int64) (*CheckinStatus, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	status := &CheckinStatus{Enabled: settings.Enabled}
	if !settings.Enabled {
		return status, nil
	}
	today := checkinToday()

	log, err := s.checkinRepo.GetByDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	if log != nil {
		status.CheckedInToday = true
		status.ConsecutiveDays = log.ConsecutiveDays
		status.TodayAmount = log.Amount
		status.LastCheckinDate = &log.CheckinDate
		return status, nil
	}

	consecutiveDays, err := s.calcConsecutiveDays(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	status.ConsecutiveDays = consecutiveDays
	status.TodayAmount = calcCheckinAmount(settings, consecutiveDays)

	last, err := s.checkinRepo.GetLastBefore(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	if last != nil {
		status.LastCheckinDate = &last.CheckinDate
	}
	return status, nil
}

func (s *CheckinService) DoCheckin(ctx context.Context, userID int64) (*CheckinResult, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, ErrCheckinDisabled
	}
	today := checkinToday()

	existing, err := s.checkinRepo.GetByDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCheckinAlreadyDone
	}

	consecutiveDays, err := s.calcConsecutiveDays(ctx, userID, today)
	if err != nil {
		return nil, err
	}
	amount := calcCheckinAmount(settings, consecutiveDays)

	if _, err = s.checkinRepo.Create(ctx, &CheckinLog{
		UserID:          userID,
		Amount:          amount,
		ConsecutiveDays: consecutiveDays,
		CheckinDate:     today,
	}); err != nil {
		return nil, err
	}

	// credit balance through AdminService so the ledger / cache invalidation stays consistent
	user, err := s.adminSvc.UpdateUserBalance(ctx, userID, amount, "add", "daily checkin reward")
	if err != nil {
		return nil, err
	}
	return &CheckinResult{Amount: amount, ConsecutiveDays: consecutiveDays, NewBalance: user.Balance}, nil
}

func (s *CheckinService) ListLogs(ctx context.Context, userID int64, page, pageSize int) ([]*CheckinLog, int64, error) {
	return s.checkinRepo.List(ctx, userID, page, pageSize)
}

func (s *CheckinService) calcConsecutiveDays(ctx context.Context, userID int64, today time.Time) (int, error) {
	last, err := s.checkinRepo.GetByDate(ctx, userID, today.AddDate(0, 0, -1))
	if err != nil {
		return 0, err
	}
	if last == nil {
		return 1, nil
	}
	return last.ConsecutiveDays + 1, nil
}

func calcCheckinAmount(s *CheckinSettings, consecutiveDays int) float64 {
	amount := s.BaseAmount
	if s.ConsecutiveBonus && consecutiveDays > 1 {
		bonus := consecutiveDays - 1
		if bonus > s.MaxBonusDays {
			bonus = s.MaxBonusDays
		}
		amount += float64(bonus) * s.BonusPerDay
	}
	return amount
}

// checkinToday returns the current UTC date at midnight, matching the Postgres date column.
func checkinToday() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func checkinParseBool(s string, def bool) bool {
	if v, err := strconv.ParseBool(s); err == nil {
		return v
	}
	return def
}

func checkinParseFloat(s string, def float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return def
}

func checkinParseInt(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
