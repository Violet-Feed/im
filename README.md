protoc -I . --go_out=. ./proto/im.proto --go-grpc_out=. ./proto/im.proto
./kvrocks/build/kvrocks -c ./kvrocks/kvrocks.conf
nohup sh ./rocketmq-all-5.3.0-source-release/distribution/target/rocketmq-5.3.0/rocketmq-5.3.0/bin/mqnamesrv &
nohup sh ./rocketmq-all-5.3.0-source-release/distribution/target/rocketmq-5.3.0/rocketmq-5.3.0/bin/mqbroker -n
localhost:9876 &
./rocketmq-all-5.3.0-source-release/distribution/target/rocketmq-5.3.0/rocketmq-5.3.0/bin/mqadmin updateTopic -n
localhost:9876 -b localhost:10911 -t user