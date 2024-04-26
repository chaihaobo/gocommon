# 介绍

本仓库为Go后端通用模块的 SDK

# 使用

## asyncContext

这个 context 用于异步任务，可以在异步任务中使用. 它将永远不会结束并且会继承父 context 上下文值.
他适用于异步开启 goroutine 传递上下文的场景

```go
ctx := http.Request{}.Context()
ctx = context.Async(ctx)
```

## logger

日志模块，使用方式如下. zap 实现. 并且会自动记录当前的 traceId.
如果 FileName 为空，则日志会标准输出和标准错误输出

```go
logger, f, err := logger.New(logger.Config{
FileName:   "",
MaxSize:    10 << 20,
MaxAge:     1, // days
WithCaller: true,
CallerSkip: 1,
})
if err != nil {
panic(err)
}
defer f()
logger.Info(context.Background(), "info message", zap.String("key", "value"))
```

## metric

指标模块. 基于open-telemetry+prometheus 实现. 使用方式如下

```go
ctx := context.Background()
meter, err := metric.NewPrometheusMetric(metric.Config{
ServiceName: "service name", // service name
Port:        0,              // 会启动 metric server 监听这个端口
Registerer:  nil, // 自定义 registerer. 不传递使用默认的实现
Gatherer:    nil, // 自定义 gatherer. 不传递使用默认的实现
})
if err != nil {
panic(err)
}

if counter, err := meter.Int64Counter("counter"); err == nil {
counter.Add(ctx, 1, otelmetric.WithAttributes(
attribute.String("status", "200"),
))
}
// or use
if counter, err := otel.Meter("meter name").Int64Counter("counter"); err == nil {
counter.Add(ctx, 1, otelmetric.WithAttributes(
attribute.String("status", "200"),
))
}

```

## trace

链路追踪模块. 基于open-telemetry + zipkin 实现.
使用方式如下

```go
tracer, err := trace.NewZipkinTracer(trace.Config{
CollectorURL: "http://localhost:9411/api/v2/spans", // 可以为空. 不会影响服务业务逻辑
ServiceName:  "service name",
})
if err != nil {
panic(err)
}
ctx := context.Background()
ctx, span := tracer.Start(ctx, "userService.login", oteltrace.WithAttributes())
defer span.End()

// or use
ctx, span = otel.Tracer("tracer name").Start(ctx, "userService.login")
defer span.End()
```

## middleware

### gin telemetry middleware

gin telemetry 中间件，用于记录请求的指标和链路追踪信息. 以及打印日志

```go
logger, f, err := logger.New(logger.Config{})
if err != nil {
panic(err)
}
defer f()
gin.New().Use(ginmiddleware.TelemetryMiddleware("service", "dev", logger))
```

