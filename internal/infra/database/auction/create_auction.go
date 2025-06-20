package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	return nil
}

func (ar *AuctionRepository) UpdateAuctionStatus(
	ctx context.Context,
	auctionId string,
	status auction_entity.AuctionStatus) *internal_error.InternalError {

	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to update auction status", err)
		return internal_error.NewInternalServerError("Error trying to update auction status")
	}

	return nil
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}

func getAuctionCheckInterval() time.Duration {
	checkInterval := os.Getenv("AUCTION_CHECK_INTERVAL")
	duration, err := time.ParseDuration(checkInterval)
	if err != nil {
		return time.Minute * 1
	}

	return duration
}

func (ar *AuctionRepository) FindActiveAuctionsToCheck(
	ctx context.Context) ([]auction_entity.Auction, *internal_error.InternalError) {

	filter := bson.M{"status": auction_entity.Active}

	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("Error finding active auctions", err)
		return nil, internal_error.NewInternalServerError("Error finding active auctions")
	}
	defer cursor.Close(ctx)

	var auctionsMongo []AuctionEntityMongo
	if err := cursor.All(ctx, &auctionsMongo); err != nil {
		logger.Error("Error decoding active auctions", err)
		return nil, internal_error.NewInternalServerError("Error decoding active auctions")
	}

	var auctionsEntity []auction_entity.Auction
	for _, auction := range auctionsMongo {
		auctionsEntity = append(auctionsEntity, auction_entity.Auction{
			Id:          auction.Id,
			ProductName: auction.ProductName,
			Category:    auction.Category,
			Status:      auction.Status,
			Description: auction.Description,
			Condition:   auction.Condition,
			Timestamp:   time.Unix(auction.Timestamp, 0),
		})
	}

	return auctionsEntity, nil
}
