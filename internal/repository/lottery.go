package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type LotteryRepository interface {
	// Event
	CreateEvent(ctx context.Context, event *entity.LotteryEvent) error
	GetEventByID(ctx context.Context, eventID string) (*entity.LotteryEvent, error)
	GetLastEventByCommunityID(ctx context.Context, communityID string) (*entity.LotteryEvent, error)
	CheckAndUseEventTicket(ctx context.Context, eventID string) error

	// Prize
	CreatePrize(ctx context.Context, prize *entity.LotteryPrize) error
	GetPrizeByID(ctx context.Context, prizeID string) (*entity.LotteryPrize, error)
	GetPrizesByIDs(ctx context.Context, prizeIDs []string) ([]entity.LotteryPrize, error)
	GetPrizesByEventID(ctx context.Context, eventID string) ([]entity.LotteryPrize, error)
	CheckAndWinEventPrize(ctx context.Context, prizeID string) error

	// Winner
	CreateWinner(ctx context.Context, winner *entity.LotteryWinner) error
	GetWinnerByID(ctx context.Context, winnerID string) (*entity.LotteryWinner, error)
	GetNotClaimedWinnerByUserID(ctx context.Context, userID string) ([]entity.LotteryWinner, error)
	ClaimWinnerReward(ctx context.Context, winnerID string) error
}

type lotteryRepository struct{}

func NewLotteryRepository() *lotteryRepository {
	return &lotteryRepository{}
}

func (r *lotteryRepository) CreateEvent(ctx context.Context, event *entity.LotteryEvent) error {
	return xcontext.DB(ctx).Create(event).Error
}

func (r *lotteryRepository) GetEventByID(ctx context.Context, eventID string) (*entity.LotteryEvent, error) {
	var result entity.LotteryEvent
	if err := xcontext.DB(ctx).Take(&result, "id=?", eventID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *lotteryRepository) GetLastEventByCommunityID(ctx context.Context, communityID string) (*entity.LotteryEvent, error) {
	var result entity.LotteryEvent
	err := xcontext.DB(ctx).Where("community_id=?", communityID).
		Order("start_time DESC").Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *lotteryRepository) CheckAndUseEventTicket(ctx context.Context, eventID string) error {
	tx := xcontext.DB(ctx).Model(&entity.LotteryEvent{}).
		Where("id=? AND used_tickets < max_tickets", eventID).
		Update("used_tickets", gorm.Expr("used_tickets+?", 1))
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *lotteryRepository) CreatePrize(ctx context.Context, prize *entity.LotteryPrize) error {
	return xcontext.DB(ctx).Create(prize).Error
}

func (r *lotteryRepository) GetPrizeByID(ctx context.Context, prizeID string) (*entity.LotteryPrize, error) {
	var result entity.LotteryPrize
	if err := xcontext.DB(ctx).Take(&result, "id=?", prizeID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *lotteryRepository) GetPrizesByIDs(ctx context.Context, prizeIDs []string) ([]entity.LotteryPrize, error) {
	var result []entity.LotteryPrize
	if err := xcontext.DB(ctx).Find(&result, "id IN (?)", prizeIDs).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *lotteryRepository) GetPrizesByEventID(ctx context.Context, eventID string) ([]entity.LotteryPrize, error) {
	var result []entity.LotteryPrize
	if err := xcontext.DB(ctx).Find(&result, "lottery_event_id=?", eventID).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *lotteryRepository) CheckAndWinEventPrize(ctx context.Context, eventRewardID string) error {
	tx := xcontext.DB(ctx).Model(&entity.LotteryPrize{}).
		Where("id=? AND won_rewards < available_rewards", eventRewardID).
		Update("won_rewards", gorm.Expr("won_rewards+?", 1))
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *lotteryRepository) CreateWinner(ctx context.Context, winner *entity.LotteryWinner) error {
	return xcontext.DB(ctx).Create(winner).Error
}

func (r *lotteryRepository) ClaimWinnerReward(ctx context.Context, winnerID string) error {
	tx := xcontext.DB(ctx).Model(&entity.LotteryWinner{}).
		Where("id=? AND is_claimed=?", winnerID, false).
		Update("is_claimed", true)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *lotteryRepository) GetWinnerByID(ctx context.Context, winnerID string) (*entity.LotteryWinner, error) {
	var result entity.LotteryWinner
	if err := xcontext.DB(ctx).Take(&result, "id=?", winnerID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *lotteryRepository) GetNotClaimedWinnerByUserID(ctx context.Context, userID string) ([]entity.LotteryWinner, error) {
	var result []entity.LotteryWinner
	if err := xcontext.DB(ctx).Find(&result, "user_id=? AND is_claimed=?", userID, false).Error; err != nil {
		return nil, err
	}

	return result, nil
}
