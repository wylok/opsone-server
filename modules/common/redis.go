package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func RedisConnect() (*redis.Client, context.Context) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	// 初始化redis连接
	ctx := context.Background()
	RC := redis.NewClient(&redis.Options{
		Addr:     cf.RedisConfig["addr"].(string),
		Password: cf.RedisConfig["password"].(string),
		DB:       0, // use default DB
	})
	err = RC.Ping(ctx).Err()
	if err != nil {
		fmt.Println(err)
		return nil, ctx
	}
	return RC, ctx
}
