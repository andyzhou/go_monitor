# monitor

Introduce
====
This is node monitor service, base on grpc, used for assigned nodes.

Dependence
====
Need go 1.17.x

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
