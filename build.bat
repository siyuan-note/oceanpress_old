SET CGO_ENABLED=1
SET GOOS=darwin
SET GOARCH=amd64
go build -x -v -ldflags "-s -w" -o mac_md2website ./src

SET CGO_ENABLED=1
SET GOOS=linux
SET GOARCH=amd64
SET CC=x86_64-linux-musl-gcc
SET CXX=aarch64-linux-gnu-g++
go build -ldflags "-linkmode external -extldflags -static" -x -v -o linux_md2website ./src

SET CGO_ENABLED=1
SET GOOS=windows
SET GOARCH=amd64
go build -x -v -o windows_md2website.exe ./src
# https://kaixuan.im/2019/12/07/macos-golang-shi-yong-sqlite3zai-m-jiao-cha-windows-liunuxbian-yi-wen-ti/