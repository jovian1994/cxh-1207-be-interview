package cache

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func TestGetRedKey(t *testing.T) {
	err := InitRedis("my-redis", &redis.Options{
		Addr: "127.0.0.1:6379",
	}, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("ok")
	}
	client := GetRedisClient("my-redis")
	client.Set("test", "test", 10*time.Second)
	data := client.GetWithContext(getContext(), "test")
	fmt.Println(data)
}
