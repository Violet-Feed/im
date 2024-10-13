protoc -I . --go_out=. ./proto/im.proto --go-grpc_out=. ./proto/im.proto
./kvrocks/build/kvrocks -c kvrocks.conf