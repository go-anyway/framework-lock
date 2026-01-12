// Copyright 2025 zampo.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// @contact  zampo3380@gmail.com

package lock

import (
	"context"
	"fmt"
	"time"
)

// RedisLock Redis 分布式锁实现
type RedisLock struct {
	client  RedisClient
	options *Options
	value   string // 锁的值（用于安全释放）
}

// NewRedisLock 创建 Redis 分布式锁
func NewRedisLock(client RedisClient, opts *Options) *RedisLock {
	if opts == nil {
		opts = &Options{
			TTL:         30 * time.Second,
			RetryDelay:  100 * time.Millisecond,
			MaxRetries:  3,
			EnableTrace: true,
		}
	}
	return &RedisLock{
		client:  client,
		options: opts,
		value:   generateLockValue(),
	}
}

// Lock 获取锁（阻塞）
func (l *RedisLock) Lock(ctx context.Context) error {
	if l.options.EnableTrace {
		return lockWithTrace(ctx, "lock", l.options.Key, func(ctx context.Context) error {
			return l.lock(ctx)
		}, true)
	}
	return l.lock(ctx)
}

// lock 内部实现
func (l *RedisLock) lock(ctx context.Context) error {
	for i := 0; i < l.options.MaxRetries; i++ {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.options.RetryDelay):
		}
	}
	return fmt.Errorf("failed to acquire lock after %d retries", l.options.MaxRetries)
}

// TryLock 尝试获取锁（非阻塞）
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	if l.options.EnableTrace {
		var ok bool
		var err error
		lockWithTrace(ctx, "try_lock", l.options.Key, func(ctx context.Context) error {
			ok, err = l.tryLock(ctx)
			return err
		}, true)
		return ok, err
	}
	return l.tryLock(ctx)
}

// tryLock 内部实现
func (l *RedisLock) tryLock(ctx context.Context) (bool, error) {
	ok, err := l.client.SetNX(ctx, l.options.Key, l.value, l.options.TTL)
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return ok, nil
}

// Unlock 释放锁
func (l *RedisLock) Unlock(ctx context.Context) error {
	if l.options.EnableTrace {
		return lockWithTrace(ctx, "unlock", l.options.Key, func(ctx context.Context) error {
			return l.unlock(ctx)
		}, true)
	}
	return l.unlock(ctx)
}

// unlock 内部实现
func (l *RedisLock) unlock(ctx context.Context) error {
	// 使用 Lua 脚本确保只释放自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := l.client.Eval(ctx, script, []string{l.options.Key}, l.value)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// Extend 延长锁的过期时间
func (l *RedisLock) Extend(ctx context.Context, duration time.Duration) error {
	if l.options.EnableTrace {
		return lockWithTrace(ctx, "extend", l.options.Key, func(ctx context.Context) error {
			return l.extend(ctx, duration)
		}, true)
	}
	return l.extend(ctx, duration)
}

// extend 内部实现
func (l *RedisLock) extend(ctx context.Context, duration time.Duration) error {
	// 使用 Lua 脚本确保只延长自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	_, err := l.client.Eval(ctx, script, []string{l.options.Key}, l.value, int(duration.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}
	return nil
}

// generateLockValue 生成锁的值（用于安全释放）
func generateLockValue() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
