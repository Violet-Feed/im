```bash
protoc -I . --go_out=. --go_opt=module=im --go-grpc_out=. --go-grpc_opt=module=im .\proto\common.proto .\proto\im.proto .\proto\action.proto .\proto\push.proto
```