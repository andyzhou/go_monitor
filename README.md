# go_monitor

Introduce
====
This is monitor service, base on grpc, used for monitor assigned nodes.

Install proto buff for go
====
go get github.com/golang/protobuf/protoc-gen-go
cd github.com/golang/protobuf/protoc-gen-go
go build
go install or `cp -f protoc-gen-go /usr/local/go/bin`

Generate pb from proto file
====
cd proto
/usr/local/bin/protoc --go_out=plugins=grpc:. *.proto

How to use?
====
see example/example.go for more details.
