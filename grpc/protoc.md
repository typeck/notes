- protoc 示例
  
  `protoc [参数] .proto文件路径`

- go文件生成:
  
  `protoc --go_out=./go/ ./proto/helloworld.proto`

- 指定import相对路径
  
  `protoc --go_out=plugins=grpc,paths=source_relative:. ./*.proto`

 `protoc -I /root/type -I ./ --go_out=plugins=grpc,paths=source_relative:. ./*.proto` 