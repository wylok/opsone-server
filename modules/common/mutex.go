package common

import (
	"errors"
	"fmt"
	"github.com/golang-module/carbon"
	"inner/conf/platform_conf"
	"time"
)

var (
	rc, ctx = RedisConnect()
)

type SyncMutex struct {
	LockKey  string
	LockTime int64
}

func (mu *SyncMutex) Lock() bool {
	// lock
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var lock bool
	if mu.LockTime == 0 {
		mu.LockTime = 30
	}
	mu.LockKey = mu.LockKey + "_" + carbon.Now().Format("YmdHi")
	if rc.Exists(ctx, mu.LockKey).Val() == 0 {
		_ = rc.Set(ctx, mu.LockKey, platform_conf.Uuid, time.Duration(mu.LockTime)*time.Second)
		lock = true
	} else {
		if rc.Get(ctx, mu.LockKey).Val() == platform_conf.Uuid {
			lock = true
		}
	}
	return lock
}
func (mu *SyncMutex) UnLock(lease bool) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	rc.Del(ctx, mu.LockKey)
	if lease {
		rc.Expire(ctx, mu.LockKey, time.Duration(mu.LockTime)*time.Second)
	}
}
