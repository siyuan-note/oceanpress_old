SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o mac_md2website ./src

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o linux_md2website ./src

SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -o windows_md2website.exe ./src
