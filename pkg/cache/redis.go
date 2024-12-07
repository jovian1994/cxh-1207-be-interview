package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/timeutil"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/trace"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var redisClients = make(map[string]ICache)

type Redis struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient
	trace         *trace.Cache
}

const (
	DefaultRedisClient = "default-redis-client"
	MinIdleConns       = 50
	PoolSize           = 20
	MaxRetries         = 3
)

type RedisMode string

const (
	RedisStandalone = "standalone"
	RedisCluster    = "cluster"
)

func setDefaultOptions(opt *redis.Options) {
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值为3秒
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值与ReadTimeout相等
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.PoolTimeout == 0 {
		//默认为ReadTimeout + 1秒（4s）
		opt.PoolTimeout = 10 * time.Second
	}
	if opt.IdleTimeout == 0 {
		//默认值为5秒
		opt.IdleTimeout = 10 * time.Second
	}
}

func setDefaultClusterOptions(opt *redis.ClusterOptions) {
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值为3秒
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.ReadTimeout == 0 {
		//默认值与ReadTimeout相等
		opt.ReadTimeout = 2 * time.Second
	}

	if opt.PoolTimeout == 0 {
		//默认为ReadTimeout + 1秒（4s）
		opt.PoolTimeout = 10 * time.Second
	}
	if opt.IdleTimeout == 0 {
		//默认值为5秒
		opt.IdleTimeout = 10 * time.Second
	}
}

func InitRedis(clientName string, opt *redis.Options, trace *trace.Cache) error {
	if len(clientName) == 0 {
		return errors.New("empty client name")
	}
	if len(opt.Addr) == 0 {
		return errors.New("empty addr")
	}
	setDefaultOptions(opt)
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis err addr: %s %s: ", err.Error(), opt.Addr)
	}
	redisClients[clientName] = &Redis{
		client: client,
		trace:  trace,
	}
	return nil
}

func InitClusterRedis(clientName string, opt *redis.ClusterOptions, trace *trace.Cache) error {
	if len(clientName) == 0 {
		return errors.New("empty client name")
	}
	if len(opt.Addrs) == 0 {
		return errors.New("empty addrs")
	}
	setDefaultClusterOptions(opt)
	//NewClusterClient执行过程中会连接redis集群并, 并尝试发送("cluster", "info")指令去进行多次连接,
	//如果这里传入很多连接地址，并且连接地址都不可用的情况下会阻塞很长时间
	client := redis.NewClusterClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis err  addrs :%s, %v", err.Error(), opt.Addrs)
	}
	redisClients[clientName] = &Redis{
		clusterClient: client,
	}
	return nil
}

func GetRedisClient(name string) ICache {
	if client, ok := redisClients[name]; ok {
		return client
	}
	return nil
}

func GetRedisClusterClient(name string) ICache {
	if client, ok := redisClients[name]; ok {
		return client
	}
	return nil
}

func getContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

// Set set some <key,value> into redis
func (r *Redis) Set(key string, value interface{}, ttl time.Duration) error {
	return r.SetWithContext(getContext(), key, value, ttl)
}

// Get get some key from redis
func (r *Redis) Get(key string) string {
	return r.GetWithContext(getContext(), key)
}

func (r *Redis) GetStr(key string) (value string, err error) {
	return r.GetStrWithContext(getContext(), key)
}

// TTL get some key from redis
func (r *Redis) TTL(key string) (time.Duration, error) {
	return r.TTLWithContext(getContext(), key)
}

// Expire expire some key
func (r *Redis) Expire(key string, ttl time.Duration) (bool, error) {
	return r.ExpireWithContext(getContext(), key, ttl)
}

// ExpireAt expire some key at some time
func (r *Redis) ExpireAt(key string, ttl time.Time) (bool, error) {
	return r.ExpireAtWithContext(getContext(), key, ttl)

}

func (r *Redis) Exists(keys ...string) (bool, error) {
	return r.ExistsWithContext(getContext(), keys...)
}

func (r *Redis) IsExist(key string) bool {
	return r.IsExistWithContext(getContext(), key)
}

func (r *Redis) Delete(key string) error {
	return r.DeleteWithContext(getContext(), key)
}

func (r *Redis) Incr(key string) (int64, error) {
	return r.IncrWithContext(getContext(), key)
}

// Close close redis client
func (r *Redis) Close() error {
	return r.client.Close()
}

// Version redis server version
func (r *Redis) Version() string {
	return r.VersionWithContext(getContext())
}

func (r *Redis) Publish(data []byte, channelName string) error {
	return r.PublishWithContext(getContext(), data, channelName)
}

func (r *Redis) Subscribe(callback func(data string), chanelNames ...string) {

	var channel = new(redis.PubSub)
	if r.client != nil {
		channel = r.client.Subscribe(context.Background(), chanelNames...)
	} else {
		channel = r.clusterClient.Subscribe(context.Background(), chanelNames...)
	}
	for msg := range channel.Channel() {
		msg := msg
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("call func error", zap.Any("error", err))
				}
			}()
			callback(msg.Payload)
		}()
	}
}

func (r *Redis) SetNX(key string, value any, expire time.Duration) (bool, error) {
	return r.SetNXWithContext(getContext(), key, value, expire)
}

func (r *Redis) BLPop(timeout time.Duration, keys ...string) ([]string, error) {
	return r.BLPopWithContext(getContext(), timeout, keys...)
}

func (r *Redis) Eval(script string, keys []string, args ...any) (any, error) {
	return r.EvalWithContext(getContext(), script, keys, args...)
}

func (r *Redis) SetWithContext(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if len(key) == 0 {
		return errors.New("empty key")
	}
	ts := time.Now()
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()
		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "set"
		r.trace.Key = key
		r.trace.Value = value
		r.trace.TTL = ttl.Minutes()
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()
	if r.client != nil {
		if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
			return fmt.Errorf("redis set key: %s err: %s", key, err.Error())
		}
		return nil
	}
	//集群版
	if err := r.clusterClient.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set key: %s err:%s", key, err.Error())
	}
	return nil
}

func (r *Redis) GetWithContext(ctx context.Context, key string) string {
	if len(key) == 0 {
		StdLogger.Println("empty key")
		return ""
	}
	ts := time.Now()
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "get"
		r.trace.Key = key
		r.trace.Value = ""
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err := r.client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			StdLogger.Printf("redis get key: %s err %v", key, err)

		}
		return value
	}

	value, err := r.clusterClient.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		StdLogger.Printf("redis get key: %s err %v", key, err)
	}
	return value
}

func (r *Redis) GetStrWithContext(ctx context.Context, key string) (value string, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "get"
		r.trace.Key = key
		r.trace.Value = value
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("redis get key: %s err:%s", key, err.Error())
		}
		return
	}
	value, err = r.clusterClient.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("redis get key: %s err: %s", key, err.Error())
	}
	return
}

func (r *Redis) TTLWithContext(ctx context.Context, key string) (time.Duration, error) {
	if len(key) == 0 {
		return 0, errors.New("empty key")
	}
	if r.client != nil {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return -1, fmt.Errorf("redis get key: %s err: %s", key, err.Error())
		}
		return ttl, nil
	}
	ttl, err := r.clusterClient.TTL(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return -1, fmt.Errorf("redis get key: %s err:%s", key, err.Error())
	}
	return ttl, nil
}

func (r *Redis) ExpireWithContext(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if len(key) == 0 {
		return false, errors.New("empty key")
	}
	if r.client != nil {
		ok, err := r.client.Expire(ctx, key, ttl).Result()
		return ok, err
	}
	ok, err := r.clusterClient.Expire(ctx, key, ttl).Result()
	return ok, err
}

func (r *Redis) ExpireAtWithContext(ctx context.Context, key string, ttl time.Time) (bool, error) {
	if len(key) == 0 {
		return false, errors.New("empty key")
	}
	if r.client != nil {
		ok, err := r.client.ExpireAt(ctx, key, ttl).Result()
		return ok, err
	}
	ok, err := r.clusterClient.ExpireAt(ctx, key, ttl).Result()
	return ok, err
}

func (r *Redis) DeleteWithContext(ctx context.Context, key string) error {
	if len(key) == 0 {
		return errors.New("empty key")
	}
	ts := time.Now()
	var value int64
	var err error
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "del"
		r.trace.Key = key
		r.trace.Value = strconv.FormatInt(value, 10)
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		_, err = r.client.Del(ctx, key).Result()
		return err
	}
	//集群版
	_, err = r.clusterClient.Del(ctx, key).Result()
	return err
}

func (r *Redis) ExistsWithContext(ctx context.Context, keys ...string) (bool, error) {
	if len(keys) == 0 {
		return false, errors.New("empty keys")
	}
	if r.client != nil {
		value, err := r.client.Exists(ctx, keys...).Result()
		return value > 0, err
	}
	value, err := r.clusterClient.Exists(ctx, keys...).Result()
	return value > 0, err
}

func (r *Redis) IsExistWithContext(ctx context.Context, key string) bool {
	if len(key) == 0 {
		return false
	}
	if r.client != nil {
		value, err := r.client.Exists(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			StdLogger.Printf("cmd : Exists ; key : %s ; err : %v", key, err)
		}
		return value > 0
	}
	value, err := r.clusterClient.Exists(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		StdLogger.Printf("cmd : Exists ; key : %s ; err : %v", key, err)
	}
	return value > 0
}

func (r *Redis) IncrWithContext(ctx context.Context, key string) (value int64, err error) {
	if len(key) == 0 {
		return 0, errors.New("empty key")
	}
	ts := time.Now()
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "Incr"
		r.trace.Key = key
		r.trace.Value = strconv.FormatInt(value, 10)
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()
	if r.client != nil {
		value, err = r.client.Incr(ctx, key).Result()
		return
	}
	value, err = r.clusterClient.Incr(ctx, key).Result()
	return
}

func (r *Redis) SetBitWithContext(ctx context.Context, key string, offset int64, val int) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()

	//为了避免过大的offset导致读取性能的问题，这里需要分桶存储
	realKey := GetKey(key, offset)
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "setbit"
		r.trace.Key = realKey
		r.trace.Value = val
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.SetBit(ctx, realKey, GetOffset(offset), val).Result()
		if err != nil {
			return value, fmt.Errorf("redis setbit key: %s err: %s", realKey, err.Error())
		}
		return
	}

	//集群版为了避免单个bitmap只会落到集群中的一个节点，这里默认对bitmap进行分捅，以平衡redis集群负载，防止单个bitmap热点问题

	value, err = r.clusterClient.SetBit(ctx, realKey, GetOffset(offset), val).Result()
	if err != nil {
		return value, fmt.Errorf("redis setbit key: %s err:%s", realKey, err.Error())
	}
	return
}

func (r *Redis) GetBitWithContext(ctx context.Context, key string, offset int64) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()
	//集群版为了避免单个bitmap只会落到集群中的一个节点，这里默认对bitmap进行分捅，以平衡redis集群负载，防止单个bitmap热点问题
	realKey := GetKey(key, offset)

	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "getbit"
		r.trace.Key = realKey
		r.trace.Value = fmt.Sprintf("origin : %d ; real: %d ", offset, GetOffset(offset))
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.GetBit(ctx, realKey, GetOffset(offset)).Result()
		if err != nil {
			return value, fmt.Errorf("redis getbit key: %s err:%s", key, err.Error())
		}
		return
	}

	value, err = r.clusterClient.GetBit(ctx, realKey, GetOffset(offset)).Result()
	if err != nil {
		return value, fmt.Errorf("redis getbit key: %s, err:%s", realKey, err.Error())
	}
	return
}

func (r *Redis) SetBigBitWithContext(ctx context.Context, key string, offset int64, val int) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()

	//为了避免过大的offset导致读取性能的问题，这里需要分桶存储
	realKey := GetBigKey(key, offset)
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "setbit"
		r.trace.Key = realKey
		r.trace.Value = val
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.SetBit(ctx, realKey, GetBigOffset(offset), val).Result()
		if err != nil {
			return value, fmt.Errorf("redis setbit key: %s err:%s", realKey, err.Error())
		}
		return
	}

	value, err = r.clusterClient.SetBit(ctx, realKey, GetBigOffset(offset), val).Result()
	if err != nil {
		return value, fmt.Errorf("redis setbit key: %s err:%s", realKey, err.Error())
	}
	return
}

func (r *Redis) GetBigBitWithContext(ctx context.Context, key string, offset int64) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()

	//为了避免过大的offset导致读取性能的问题，这里需要分桶存储
	realKey := GetKey(key, offset)
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "getbit"
		r.trace.Key = realKey
		r.trace.Value = fmt.Sprintf("origin : %d ; real: %d ", offset, GetBigOffset(offset))
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.GetBit(ctx, realKey, GetOffset(offset)).Result()
		if err != nil {
			return value, fmt.Errorf("redis getbit key: %s err:%s", realKey, err.Error())
		}
		return
	}
	//集群版为了避免单个bitmap只会落到集群中的一个节点，这里默认对bitmap进行分捅，以平衡redis集群负载，防止单个bitmap热点问题
	//对于超过redis bitmap范围的数据，采用不同的分捅方式
	value, err = r.clusterClient.GetBit(ctx, realKey, GetBigOffset(offset)).Result()
	if err != nil {
		return value, fmt.Errorf("redis getbit key: %s err:%s", realKey, err.Error())
	}
	return
}

func (r *Redis) SetBitNOBucketWithContext(ctx context.Context, key string, offset int64, val int) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()

	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "setbit"
		r.trace.Key = key
		r.trace.Value = val
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.SetBit(ctx, key, offset, val).Result()
		if err != nil {
			return value, fmt.Errorf("redis setbit key: %s err:%s", key, err.Error())
		}
		return
	}

	value, err = r.clusterClient.SetBit(ctx, key, offset, val).Result()
	if err != nil {
		return value, fmt.Errorf("redis setbit key: %s err:%s", key, err.Error())
	}
	return
}

func (r *Redis) GetBitNOBucketWithContext(ctx context.Context, key string, offset int64) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "getbit"
		r.trace.Key = key
		r.trace.Value = offset
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.GetBit(ctx, key, offset).Result()
		if err != nil {
			return value, fmt.Errorf("redis getbit key: %s err:%s", key, err.Error())
		}
		return
	}

	value, err = r.clusterClient.GetBit(ctx, key, offset).Result()
	if err != nil {
		return value, fmt.Errorf("redis getbit key: %s err:%s", key, err.Error())
	}
	return
}

func (r *Redis) BitCountNOBucketWithContext(ctx context.Context, key string, start, end int64) (value int64, err error) {
	if len(key) == 0 {
		err = errors.New("empty key")
		return
	}
	ts := time.Now()
	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "bitcount"
		r.trace.Key = key
		r.trace.Value = fmt.Sprintf("start : %d ; end : %d", start, end)
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	if r.client != nil {
		value, err = r.client.BitCount(ctx, key, &redis.BitCount{
			Start: start,
			End:   end,
		}).Result()
		if err != nil {
			return value, fmt.Errorf("redis bitcount key: %s err: %s", key, err.Error())
		}
		return
	}
	value, err = r.clusterClient.BitCount(ctx, key, &redis.BitCount{
		Start: start,
		End:   end,
	}).Result()
	if err != nil {
		return value, fmt.Errorf("redis bitcount key: %s err:%s", key, err.Error())
	}
	return
}

func (r *Redis) PublishWithContext(ctx context.Context, data []byte, channelName string) error {
	if r.client != nil {
		_, err := r.client.Publish(ctx, channelName, data).Result()
		return err
	}
	_, err := r.clusterClient.Publish(ctx, channelName, data).Result()
	return err
}

func (r *Redis) SubscribeWithContext(ctx context.Context, callback func(data string), chanelNames ...string) {
	var channel = new(redis.PubSub)
	if r.client != nil {
		channel = r.client.Subscribe(ctx, chanelNames...)
	} else {
		channel = r.clusterClient.Subscribe(ctx, chanelNames...)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-channel.Channel():
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("call func error", zap.Any("error", err))
					}
				}()
				callback(msg.Payload)
			}()
		}
	}
}

func (r *Redis) SetNXWithContext(ctx context.Context, key string, value any, expire time.Duration) (bool, error) {
	if r.client != nil {
		return r.client.SetNX(ctx, key, value, expire).Result()
	}
	return r.clusterClient.SetNX(ctx, key, value, expire).Result()
}

func (r *Redis) BLPopWithContext(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	if r.client != nil {
		return r.client.BLPop(ctx, timeout, keys...).Result()
	}
	return r.clusterClient.BLPop(ctx, timeout, keys...).Result()
}

func (r *Redis) EvalWithContext(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	if r.client != nil {
		return r.client.Eval(ctx, script, keys, args...).Result()
	}
	return r.clusterClient.Eval(ctx, script, keys, args...).Result()
}

func (r *Redis) VersionWithContext(ctx context.Context) string {
	if r.client != nil {
		server := r.client.Info(ctx, "server").Val()
		spl1 := strings.Split(server, "# Server")
		spl2 := strings.Split(spl1[1], "redis_version:")
		spl3 := strings.Split(spl2[1], "redis_git_sha1:")
		return spl3[0]
	}
	server := r.clusterClient.Info(ctx, "server").Val()
	spl1 := strings.Split(server, "# Server")
	spl2 := strings.Split(spl1[1], "redis_version:")
	spl3 := strings.Split(spl2[1], "redis_git_sha1:")
	return spl3[0]
}

//对于超出redis bitmap范围的数据我们使用高49位作捅，低15为作offset

// 高49位作捅，低15为作offset
func GetBigBucket(ID int64) int64 {
	return ID >> 15
}

// 0x7FFF的二进制为111111111111111
// 与ID做与运算结果保留了ID的低15位
func GetBigOffset(ID int64) int64 {
	return ID & 0x7FFF
}

// 对于redis bitmap范围内的数据，使用高16位作捅，低16位作offset
func GetBucket(userID int64) int64 {
	return userID >> 16
}

func GetOffset(ID int64) int64 {
	return ID & 0xFFFF
}

func GetKey(key string, ID int64) string {
	return fmt.Sprintf("%s_%d", key, GetBucket(ID))
}

func GetBigKey(key string, ID int64) string {
	return fmt.Sprintf("%s_%d", key, GetBigBucket(ID))
}

func (r *Redis) GetBit(key string, offset int64) (value int64, err error) {
	return r.GetBitWithContext(context.Background(), key, offset)
}

func (r *Redis) GetBigBit(key string, offset int64) (value int64, err error) {
	return r.GetBigBitWithContext(context.Background(), key, offset)
}

func (r *Redis) SetBit(key string, offset int64, val int) (value int64, err error) {
	return r.SetBitWithContext(getContext(), key, offset, val)
}

func (r *Redis) SetBigBit(key string, offset int64, val int) (value int64, err error) {
	return r.SetBigBitWithContext(getContext(), key, offset, val)
}

func (r *Redis) GetBitNOBucket(key string, offset int64) (value int64, err error) {
	return r.GetBitNOBucketWithContext(getContext(), key, offset)
}

func (r *Redis) BitCountNOBucket(key string, start, end int64) (value int64, err error) {
	return r.BitCountNOBucketWithContext(getContext(), key, start, end)
}

func (r *Redis) SetBitNOBucket(key string, offset int64, val int) (value int64, err error) {
	return r.SetBitNOBucketWithContext(getContext(), key, offset, val)
}

func (r *Redis) BitOPNOBucket(op, destKey string, keys ...string) (value int64, err error) {
	if len(keys) == 0 {
		err = errors.New("empty keys")
		return
	}
	ts := time.Now()

	defer func() {
		if r.trace == nil || r.trace.Logger == nil {
			return
		}
		costMillisecond := time.Since(ts).Milliseconds()

		if !r.trace.AlwaysTrace && costMillisecond < r.trace.SlowLoggerMillisecond {
			return
		}
		r.trace.TraceTime = timeutil.CSTLayoutString()
		r.trace.CMD = "bitop " + op
		r.trace.Key = destKey
		r.trace.Value = strings.Join(keys, ",")
		r.trace.CostMillisecond = costMillisecond
		r.trace.Logger.Warn("redis-trace", zap.Any("", r.trace))
	}()

	var cmd *redis.IntCmd
	op = strings.ToUpper(op)
	if r.client != nil {
		switch op {
		case "AND":
			cmd = r.client.BitOpAnd(getContext(), destKey, keys...)
		case "OR":
			cmd = r.client.BitOpOr(getContext(), destKey, keys...)
		case "XOR":
			cmd = r.client.BitOpXor(getContext(), destKey, keys...)
		case "NOT":
			cmd = r.client.BitOpNot(getContext(), destKey, keys[0])
		default:
			return 0, errors.New("illegal op " + op + "; key: " + destKey)
		}
		value, err = cmd.Result()
		if err != nil {
			return value, fmt.Errorf("redis bitop AND destKey: %s  keys : %v ,err:%s", destKey, keys, err.Error())
		}
		return
	}
	switch op {
	case "AND":
		cmd = r.client.BitOpAnd(getContext(), destKey, keys...)
	case "OR":
		cmd = r.client.BitOpOr(getContext(), destKey, keys...)
	case "XOR":
		cmd = r.client.BitOpXor(getContext(), destKey, keys...)
	case "NOT":
		cmd = r.client.BitOpNot(getContext(), destKey, keys[0])
	default:
		return 0, errors.New("illegal op " + op + "; key: " + destKey)
	}
	value, err = cmd.Result()
	if err != nil {
		return value, fmt.Errorf("redis bitop %s destKey: %s  keys : %v ,err:%s", op, destKey, keys, err.Error())
	}
	return
}
