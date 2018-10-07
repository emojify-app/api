package emojify

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-redis/redis"
)

// Cache defines an interface for an image cache
type Cache interface {
	// Exists checks if an item exists in the cache
	Exists(string) (bool, error)
	// Get an image from the cache, returns true, image if found in cache or false, nil if image not found
	Get(string) ([]byte, error)
	// Put an image into the cache, returns an error if unsuccessful
	Put(string, []byte) error
}

// RedisCache implements the Cache API and uses Redis as a backend
type RedisCache struct {
	client     *redis.Client
	expiration time.Duration
}

// NewRedisCache creates a new RedisCache with the given connection string
func NewRedisCache(connection string) Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     connection,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &RedisCache{client, 120 * time.Second}
}

// Exists checks to see if a key exists in the cache
func (r *RedisCache) Exists(key string) (bool, error) {
	c, err := r.client.Exists(sanitiseKey(key)).Result()
	return (c > 0), err
}

// Get an image from the Redis store
func (r *RedisCache) Get(key string) ([]byte, error) {
	return r.client.Get(key).Bytes()
}

// Put an image to the Redis store
func (r *RedisCache) Put(key string, data []byte) error {
	return r.client.Set(key, data, r.expiration).Err()
}

// FileCache implements the Cache interface using the local filesystem
type FileCache struct {
	path string
}

// NewFileCache creates a file based cache
func NewFileCache(path string) Cache {
	return &FileCache{path}
}

// Exists checks to see if a file
func (r *FileCache) Exists(key string) (bool, error) {
	_, err := os.Open(r.path + sanitiseKey(key))
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

// Get an image from the File store
func (r *FileCache) Get(key string) ([]byte, error) {
	f, err := os.Open(r.path + sanitiseKey(key))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

// Put an image to the File store
func (r *FileCache) Put(key string, data []byte) error {
	f, err := os.Create(r.path + sanitiseKey(key))
	if err != nil {
		return err
	}

	_, err = f.Write(data)

	return err
}

// encodes the key replacing any non-suitable characters
func sanitiseKey(key string) string {
	return base64.StdEncoding.EncodeToString([]byte(key))
}
