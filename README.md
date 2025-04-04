protoc -I . --go_out=. ./proto/common.proto --go-grpc_out=. ./proto/common.proto
protoc -I . --go_out=. ./proto/im.proto --go-grpc_out=. ./proto/im.proto
protoc -I . --go_out=. ./proto/action.proto --go-grpc_out=. ./proto/action.proto
protoc -I . --go_out=. ./proto/push.proto --go-grpc_out=. ./proto/push.proto
修改proto_gen文件的import路径