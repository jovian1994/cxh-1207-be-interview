package cache

import (
	"context"
	"log"
	"os"
	"time"
)

type ICache interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) string
	GetStr(key string) (value string, err error)
	TTL(key string) (time.Duration, error)
	Expire(key string, ttl time.Duration) (bool, error)
	ExpireAt(key string, ttl time.Time) (bool, error)
	Delete(key string) error
	Exists(keys ...string) (bool, error)
	IsExist(key string) bool
	Incr(key string) (int64, error)
	SetBit(key string, offset int64, val int) (value int64, err error)
	GetBit(key string, offset int64) (value int64, err error)
	SetBigBit(key string, offset int64, val int) (value int64, err error)
	GetBigBit(key string, offset int64) (value int64, err error)
	SetBitNOBucket(key string, offset int64, val int) (value int64, err error)
	GetBitNOBucket(key string, offset int64) (value int64, err error)
	BitCountNOBucket(key string, start, end int64) (value int64, err error)
	Publish(data []byte, channelName string) error
	Subscribe(callback func(data string), chanelNames ...string)
	SetNX(key string, value any, expire time.Duration) (bool, error)
	BLPop(timeout time.Duration, keys ...string) ([]string, error)
	Eval(script string, keys []string, args ...any) (any, error)
	Close() error
	Version() string

	SetWithContext(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetWithContext(ctx context.Context, key string) string
	GetStrWithContext(ctx context.Context, key string) (value string, err error)
	TTLWithContext(ctx context.Context, key string) (time.Duration, error)
	ExpireWithContext(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ExpireAtWithContext(ctx context.Context, key string, ttl time.Time) (bool, error)
	DeleteWithContext(ctx context.Context, key string) error
	ExistsWithContext(ctx context.Context, keys ...string) (bool, error)
	IsExistWithContext(ctx context.Context, key string) bool

	IncrWithContext(ctx context.Context, key string) (int64, error)
	SetBitWithContext(ctx context.Context, key string, offset int64, val int) (value int64, err error)
	GetBitWithContext(ctx context.Context, key string, offset int64) (value int64, err error)
	SetBigBitWithContext(ctx context.Context, key string, offset int64, val int) (value int64, err error)
	GetBigBitWithContext(ctx context.Context, key string, offset int64) (value int64, err error)
	SetBitNOBucketWithContext(ctx context.Context, key string, offset int64, val int) (value int64, err error)
	GetBitNOBucketWithContext(ctx context.Context, key string, offset int64) (value int64, err error)
	BitCountNOBucketWithContext(ctx context.Context, key string, start, end int64) (value int64, err error)
	PublishWithContext(ctx context.Context, data []byte, channelName string) error
	SubscribeWithContext(ctx context.Context, callback func(data string), chanelNames ...string)
	SetNXWithContext(ctx context.Context, key string, value any, expire time.Duration) (bool, error)
	BLPopWithContext(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error)
	EvalWithContext(ctx context.Context, script string, keys []string, args ...any) (any, error)
	VersionWithContext(ctx context.Context) string
}

type stdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

var StdLogger stdLogger

func init() {
	StdLogger = log.New(os.Stdout, "[Cache] ", log.LstdFlags|log.Lshortfile)
}
