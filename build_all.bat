go get github.com/mastahyeti/cms
go get github.com/lxn/walk
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go generate
go build -ldflags="-H windowsgui" -o p7sExtract.exe
