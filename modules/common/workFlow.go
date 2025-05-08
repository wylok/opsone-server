package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
)

// DFlow 分布式workflow
type DFlow struct {
	RC        *redis.Client
	Ctx       context.Context
	LockKey   string
	Func      map[string]Func
	Depend    map[string][]string
	FuncForce map[string]bool
	Force     bool
}

// workflow引擎
type flowCore struct {
	fu map[string]*flowStruct
}

type Func func(interface{}) (interface{}, error)

type flowStruct struct {
	Deps  []string
	Ctr   int
	Fn    Func
	C     chan error
	Res   interface{}
	force bool //是否强制
	once  sync.Once
}

// workflow节点已执行
func (fs *flowStruct) done(e error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	for i := 0; i < fs.Ctr; i++ {
		fs.C <- e
	}
}

// 关闭workflow节点channel
func (fs *flowStruct) close() {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	fs.once.Do(func() {
		close(fs.C)
	})
}

// 创建workflow
func create() *flowCore {
	return &flowCore{
		fu: make(map[string]*flowStruct),
	}
}

// 增加workflow节点
func (flw *flowCore) add(name string, d []string, fn Func, fc bool) *flowCore {
	flw.fu[name] = &flowStruct{
		Deps:  d,
		Fn:    fn,
		Ctr:   1,
		force: fc,
	}
	return flw
}

// 执行workflow节点
func (flw *flowCore) start(ctx context.Context) map[string]error {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	result := map[string]error{}
	for name, fn := range flw.fu {
		for _, dep := range fn.Deps {
			// prevent self depends
			if dep == name {
				return map[string]error{name: errors.New(name + " not depends of it self")}
			}
			// prevent no existing dependencies
			if _, exists := flw.fu[dep]; exists == false {
				return map[string]error{name: errors.New(dep + " not exists")}
			}
			flw.fu[dep].Ctr++
		}
	}
	for name, fs := range flw.fu {
		fs.C = make(chan error)
		func(ctx context.Context, name string, fs *flowStruct) {
			do := true
			defer func() {
				if r := recover(); r != nil {
					fmt.Println(r)
				}
				select {
				case <-ctx.Done():
					fs.close()
				}
			}()
			if len(fs.Deps) > 0 {
				for _, dep := range fs.Deps {
					err, ok := <-flw.fu[dep].C
					if !fs.force && (err != nil || !ok) {
						do = false
					}
				}
			}
			if do {
				//匹配pipeline条件
				if len(fs.Deps) == 1 {
					fs.Res, err = fs.Fn(flw.fu[fs.Deps[0]].Res)
					result[name] = err
				} else {
					fs.Res, err = fs.Fn(nil)
					result[name] = err
				}
				fs.done(result[name])
			}
		}(ctx, name, fs)
	}
	return result
}

// Run workflow
func (df *DFlow) Run() map[string]error {
	lock := SyncMutex{LockKey: df.LockKey}
	//加锁
	if lock.Lock() {
		defer func() {
			// 释放锁
			lock.UnLock(true)
		}()
		defer func() {
			if r := recover(); r != nil {
				Log.Error(errors.New(fmt.Sprint(r)))
			}
		}()
		var force bool
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		fl := create()
		for k, v := range df.Depend {
			//默认使用全局配置
			force = df.Force
			if df.FuncForce != nil {
				_, ok := df.FuncForce[k]
				if ok {
					// 单独配置优先
					force = df.FuncForce[k]
				}
			}
			fl.add(k, v, df.Func[k], force)
		}
		return fl.start(ctx)
	}
	return nil
}
