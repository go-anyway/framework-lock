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
	"time"
)

// Lock 分布式锁接口
type Lock interface {
	// Lock 获取锁
	Lock(ctx context.Context) error
	// Unlock 释放锁
	Unlock(ctx context.Context) error
	// TryLock 尝试获取锁（非阻塞）
	TryLock(ctx context.Context) (bool, error)
	// Extend 延长锁的过期时间
	Extend(ctx context.Context, duration time.Duration) error
}

// RedisClient Redis 客户端接口
type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) error
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}
