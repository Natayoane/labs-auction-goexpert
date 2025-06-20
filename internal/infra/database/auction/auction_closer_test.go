package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"testing"
	"time"
)

type MockAuctionRepository struct {
	auctions map[string]*auction_entity.Auction
}

func NewMockAuctionRepository() *MockAuctionRepository {
	return &MockAuctionRepository{
		auctions: make(map[string]*auction_entity.Auction),
	}
}

func (m *MockAuctionRepository) CreateAuction(ctx context.Context, auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	m.auctions[auctionEntity.Id] = auctionEntity
	return nil
}

func (m *MockAuctionRepository) FindAuctions(ctx context.Context, status auction_entity.AuctionStatus, category, productName string) ([]auction_entity.Auction, *internal_error.InternalError) {
	var result []auction_entity.Auction
	for _, auction := range m.auctions {
		if status == 0 || auction.Status == status {
			result = append(result, *auction)
		}
	}
	return result, nil
}

func (m *MockAuctionRepository) FindAuctionById(ctx context.Context, id string) (*auction_entity.Auction, *internal_error.InternalError) {
	if auction, exists := m.auctions[id]; exists {
		return auction, nil
	}
	return nil, internal_error.NewNotFoundError("Auction not found")
}

func (m *MockAuctionRepository) UpdateAuctionStatus(ctx context.Context, auctionId string, status auction_entity.AuctionStatus) *internal_error.InternalError {
	if auction, exists := m.auctions[auctionId]; exists {
		auction.Status = status
		return nil
	}
	return internal_error.NewNotFoundError("Auction not found")
}

func (m *MockAuctionRepository) FindActiveAuctionsToCheck(ctx context.Context) ([]auction_entity.Auction, *internal_error.InternalError) {
	var result []auction_entity.Auction
	for _, auction := range m.auctions {
		if auction.Status == auction_entity.Active {
			result = append(result, *auction)
		}
	}
	return result, nil
}

func TestAuctionCloser_CheckAndCloseExpiredAuctions(t *testing.T) {
	mockRepo := NewMockAuctionRepository()

	expiredAuction := &auction_entity.Auction{
		Id:          "test-auction-1",
		ProductName: "Test Product",
		Category:    "Electronics",
		Description: "A test product for auction",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(-2 * time.Second),
	}
	mockRepo.CreateAuction(context.Background(), expiredAuction)

	activeAuction := &auction_entity.Auction{
		Id:          "test-auction-2",
		ProductName: "Test Product 2",
		Category:    "Electronics",
		Description: "Another test product for auction",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(-500 * time.Millisecond),
	}
	mockRepo.CreateAuction(context.Background(), activeAuction)

	// Cria o fechador de leilões
	closer := &AuctionCloser{
		auctionRepository: mockRepo,
		checkInterval:     500 * time.Millisecond,
		auctionInterval:   1 * time.Second,
		stopChan:          make(chan struct{}),
	}

	closer.checkAndCloseExpiredAuctions(context.Background())

	expiredAuctionAfter, err := mockRepo.FindAuctionById(context.Background(), "test-auction-1")
	if err != nil {
		t.Fatalf("Failed to find expired auction: %v", err)
	}

	if expiredAuctionAfter.Status != auction_entity.Completed {
		t.Errorf("Expected expired auction to be completed, got status: %d", expiredAuctionAfter.Status)
	}

	activeAuctionAfter, err := mockRepo.FindAuctionById(context.Background(), "test-auction-2")
	if err != nil {
		t.Fatalf("Failed to find active auction: %v", err)
	}

	if activeAuctionAfter.Status != auction_entity.Active {
		t.Errorf("Expected active auction to remain active, got status: %d", activeAuctionAfter.Status)
	}
}

func TestAuctionCloser_StartAndStop(t *testing.T) {
	mockRepo := NewMockAuctionRepository()

	closer := &AuctionCloser{
		auctionRepository: mockRepo,
		checkInterval:     100 * time.Millisecond,
		auctionInterval:   1 * time.Second,
		stopChan:          make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	closer.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	closer.Stop()

	t.Log("Auction closer started and stopped successfully")
}
