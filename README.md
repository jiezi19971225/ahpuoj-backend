## Build
```
go build mian.go
```

## GORM使用注意

### 主从
强制走主库 `Clauses(dbresolver.Write)`

### 更新
默认的结构体 update，如果结构体字段值是类型的默认值，比如 int 型字段值是 0 时，gorm会默认忽略，导致更新与预期不符， 需要手动Select指定要更新的字段。