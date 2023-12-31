# mgorm
帮助你简单管理 gorm 连接，多数据库连接，Sass 多数据库切换

```shell
go mod init github.com/qq1060656096/jjmgorm
```

```shell
# 代码静态检查发现可能的bug或者可疑的构造
go vet .

# 竞态检测
go build -race -v .

# 开启本地官网
go get -v -u golang.org/x/tools/cmd/godoc
godoc -http=:8080 
```

```shell
# 运行所有单元测试
# 在当前目录运行所有测试，并且每个测试函数都会执行一次，同时输出详细的测试信息。这对于调试和查看测试详细输出很有用。
go test -count=1 -v . 

# 运行所有单元测试，并查看测试覆盖率
go test -count=1 -v -cover .

# 运行所有单元测试，并查看测试覆盖率，竞态检测
go test -count=1 -v -cover -race .
```