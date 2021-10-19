##  dial-mysql-via-ssh
Golang 通过ssh代理连接mysql数据库的示例代码.

### 依赖:
- [go-sql-driver](https://github.com/go-sql-driver/mysql) 
- [ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) 
- [Viper](https://github.com/spf13/viper)
- [Cobra](https://github.com/spf13/cobra)

### 运行
```shell
$ go mod tidy

# tip: 需要配置ssh,mysql连接信息
$ go run ssh-mysql.go
```




