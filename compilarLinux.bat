set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -a -tags netgo -o srvimgdocker main.go