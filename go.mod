module github.com/siyuan-note/oceanpress

go 1.15

require (
	github.com/88250/gulu v1.1.73 // indirect
	github.com/88250/lute v1.7.4-0.20210818085620-40fd2957c84d // https://pkg.go.dev/github.com/88250/lute
	github.com/88250/protyle v0.0.0-20210807012712-9e3fb38884a3
	github.com/alecthomas/chroma v0.9.2
	github.com/alecthomas/repr v0.0.0-20210801044451-80ca428c5142 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.8
	github.com/otiai10/copy v1.6.0
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// replace github.com/88250/lute => D:\view_code\lute
// 配置国内代理 go env -w GOPROXY=https://goproxy.cn,direct
// 更新  go get github.com/88250/lute@master
// 更新  go get github.com/88250/protyle@main
