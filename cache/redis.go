package cache

import (
	"context"
	"encoding/json"
	"time"

	"example.com/cloudinary-proxy/models"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client


func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // WSL Redis address
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
}

func CacheArtworks(ctx context.Context, key string, artworks []models.Artwork, expiration time.Duration) error {
	serialized, err := json.Marshal(artworks)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, serialized, expiration).Err()
}

func GetCachedArtworks(ctx context.Context, key string) ([]models.Artwork, error) {
	val, err := RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var artworks []models.Artwork
	if err := json.Unmarshal(val, &artworks); err != nil {
		return nil, err
	}
	return artworks, nil
}