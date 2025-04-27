@echo off
echo Generating protobuf files...

cd %~dp0..

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/product.proto

echo Done!
