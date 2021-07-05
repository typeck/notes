# 表达式
## qps
```
sum(rate(req_count{cluster="$cluster", app="$app", kubernetes_namespace="ads"))
```

## sum
```
sum(increase(error_count{cluster="$cluster", app="$app", kubernetes_namespace="ads"}[5m]))

```

## p99
```
histogram_quantile(0.99, sum(rate(process_time_bucket{cluster="$cluster",app="$app"}[5m])) by (le))
```

## go gc
```
increase(go_gc_duration_seconds_count{cluster="$cluster", app="$app", kubernetes_namespace="ads"}[5m])
```
## goroutines
```
go_goroutines{cluster="$cluster", app="$app", kubernetes_namespace="ads"}
```
## 耗时百分比
```
sum(increase(_process_time_bucket{cluster="$cluster", app="$app", kubernetes_namespace="ads", le="10"}[5m]))/sum(increase(_process_time_bucket{cluster="$cluster", app="$app", kubernetes_namespace="ads", le="+Inf"}[5m]))
```

## qps同比
```
sum(rate(_req_count{cluster="$cluster", app="$app", kubernetes_namespace="ads", click_type="client"}[20m]))/sum(rate(req_count{cluster="$cluster", app="$app", kubernetes_namespace="ads"click_type="client"}[20m] offset 1d))
```
## 平均耗时
```
sum(cost_sum{cluster="$cluster"}) / sum(cost_count{cluster="$cluster"})
```
