package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"sync"
	"time"

	"go.uber.org/zap"
)

type AuctionCloser struct {
	auctionRepository auction_entity.AuctionRepositoryInterface
	checkInterval     time.Duration
	auctionInterval   time.Duration
	stopChan          chan struct{}
	wg                sync.WaitGroup
}

func NewAuctionCloser(auctionRepository auction_entity.AuctionRepositoryInterface) *AuctionCloser {
	return &AuctionCloser{
		auctionRepository: auctionRepository,
		checkInterval:     getAuctionCheckInterval(),
		auctionInterval:   getAuctionInterval(),
		stopChan:          make(chan struct{}),
	}
}

func (ac *AuctionCloser) Start(ctx context.Context) {
	ac.wg.Add(1)
	go ac.run(ctx)
}

func (ac *AuctionCloser) Stop() {
	close(ac.stopChan)
	ac.wg.Wait()
}

func (ac *AuctionCloser) run(ctx context.Context) {
	defer ac.wg.Done()

	ticker := time.NewTicker(ac.checkInterval)
	defer ticker.Stop()

	logger.Info("Auction closer routine started",
		zap.String("check_interval", ac.checkInterval.String()),
		zap.String("auction_interval", ac.auctionInterval.String()),
	)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Auction closer routine stopped due to context cancellation")
			return
		case <-ac.stopChan:
			logger.Info("Auction closer routine stopped")
			return
		case <-ticker.C:
			ac.checkAndCloseExpiredAuctions(ctx)
		}
	}
}

func (ac *AuctionCloser) checkAndCloseExpiredAuctions(ctx context.Context) {
	activeAuctions, err := ac.auctionRepository.FindActiveAuctionsToCheck(ctx)
	if err != nil {
		logger.Error("Error finding active auctions to check", err)
		return
	}

	now := time.Now()
	closedCount := 0

	for _, auction := range activeAuctions {
		expirationTime := auction.Timestamp.Add(ac.auctionInterval)

		if now.After(expirationTime) {
			if err := ac.auctionRepository.UpdateAuctionStatus(ctx, auction.Id, auction_entity.Completed); err != nil {
				logger.Error("Error closing expired auction", err,
					zap.String("auction_id", auction.Id),
				)
				continue
			}

			logger.Info("Auction closed automatically",
				zap.String("auction_id", auction.Id),
				zap.String("product_name", auction.ProductName),
				zap.Time("created_at", auction.Timestamp),
				zap.Time("expired_at", expirationTime),
				zap.Time("closed_at", now),
			)

			closedCount++
		}
	}

	if closedCount > 0 {
		logger.Info("Auction closer processed expired auctions",
			zap.Int("closed_count", closedCount),
			zap.Int("total_checked", len(activeAuctions)),
		)
	}
}
