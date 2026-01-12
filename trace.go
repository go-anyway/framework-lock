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

	"github.com/go-anyway/framework-log"
	pkgtrace "github.com/go-anyway/framework-trace"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// lockWithTrace 带追踪的锁操作包装器
func lockWithTrace(
	ctx context.Context,
	operation string,
	key string,
	handler func(context.Context) error,
	enableTrace bool,
) error {
	startTime := time.Now()

	// 创建追踪 span
	var span trace.Span
	if enableTrace {
		ctx, span = pkgtrace.StartSpan(ctx, "lock."+operation,
			trace.WithAttributes(
				attribute.String("lock.operation", operation),
				attribute.String("lock.key", key),
			),
		)
		defer span.End()
	}

	// 执行操作
	err := handler(ctx)
	duration := time.Since(startTime)

	// 记录日志
	if err != nil {
		log.FromContext(ctx).Error("Lock operation failed",
			zap.String("operation", operation),
			zap.String("key", key),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		log.FromContext(ctx).Info("Lock operation completed",
			zap.String("operation", operation),
			zap.String("key", key),
			zap.Duration("duration", duration),
		)
	}

	// 更新追踪状态
	if enableTrace && span != nil {
		span.SetAttributes(
			attribute.Float64("lock.duration_ms", float64(duration.Milliseconds())),
		)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}

	return err
}

// tryLockWithTrace 带追踪的尝试获取锁包装器（返回 bool, error）
func tryLockWithTrace(
	ctx context.Context,
	key string,
	handler func(context.Context) (bool, error),
	enableTrace bool,
) (bool, error) {
	startTime := time.Now()

	// 创建追踪 span
	var span trace.Span
	if enableTrace {
		ctx, span = pkgtrace.StartSpan(ctx, "lock.try_lock",
			trace.WithAttributes(
				attribute.String("lock.operation", "try_lock"),
				attribute.String("lock.key", key),
			),
		)
		defer span.End()
	}

	// 执行操作
	ok, err := handler(ctx)
	duration := time.Since(startTime)

	// 记录日志
	if err != nil {
		log.FromContext(ctx).Error("Lock try lock failed",
			zap.String("key", key),
			zap.Bool("acquired", ok),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		log.FromContext(ctx).Info("Lock try lock completed",
			zap.String("key", key),
			zap.Bool("acquired", ok),
			zap.Duration("duration", duration),
		)
	}

	// 更新追踪状态
	if enableTrace && span != nil {
		span.SetAttributes(
			attribute.Bool("lock.acquired", ok),
			attribute.Float64("lock.duration_ms", float64(duration.Milliseconds())),
		)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}

	return ok, err
}
