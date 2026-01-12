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
	"fmt"
	"time"

	pkgConfig "github.com/go-anyway/framework-config"
)

// Config 分布式锁配置结构体（用于从配置文件创建）
type Config struct {
	Enabled     bool               `yaml:"enabled" env:"LOCK_ENABLED" default:"true"`
	TTL         pkgConfig.Duration `yaml:"ttl" env:"LOCK_TTL" default:"30s"`
	RetryDelay  pkgConfig.Duration `yaml:"retry_delay" env:"LOCK_RETRY_DELAY" default:"100ms"`
	MaxRetries  int                `yaml:"max_retries" env:"LOCK_MAX_RETRIES" default:"3"`
	EnableTrace bool               `yaml:"enable_trace" env:"LOCK_ENABLE_TRACE" default:"true"`
}

// Validate 验证分布式锁配置
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("lock config cannot be nil")
	}
	if !c.Enabled {
		return nil // 如果未启用，不需要验证
	}
	if c.TTL.Duration() <= 0 {
		return fmt.Errorf("lock ttl must be greater than 0")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("lock max_retries must be non-negative")
	}
	return nil
}

// ToOptions 转换为 Options
func (c *Config) ToOptions() (*Options, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if !c.Enabled {
		return nil, fmt.Errorf("lock is not enabled")
	}

	ttl := c.TTL.Duration()
	if ttl == 0 {
		ttl = 30 * time.Second
	}

	retryDelay := c.RetryDelay.Duration()
	if retryDelay == 0 {
		retryDelay = 100 * time.Millisecond
	}

	maxRetries := c.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &Options{
		TTL:         ttl,
		RetryDelay:  retryDelay,
		MaxRetries:  maxRetries,
		EnableTrace: c.EnableTrace,
	}, nil
}

// Options 分布式锁配置选项（内部使用）
type Options struct {
	Key         string        // 锁的键
	TTL         time.Duration // 锁的过期时间
	RetryDelay  time.Duration // 重试延迟
	MaxRetries  int           // 最大重试次数
	EnableTrace bool          // 是否启用追踪
}
