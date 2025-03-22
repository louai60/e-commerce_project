@echo off

:: Ensure GOPATH\bin is in PATH for protoc plugins
set PATH=%PATH%;%GOPATH%\bin

:: Create proto directory if it doesn't exist
if not exist proto mkdir proto

:: Show protoc version and plugin locations for debugging
echo Protoc version:
protoc --version
echo.
echo Checking protoc plugins...
where protoc-gen-go
where protoc-gen-go-grpc
echo.

:: Generate the protobuf and gRPC code
echo Generating protobuf and gRPC code...
protoc --proto_path=. ^
       --go_out=. ^
       --go_opt=paths=source_relative ^
       --go-grpc_out=. ^
       --go-grpc_opt=paths=source_relative ^
       proto\product.proto

if %ERRORLEVEL% EQU 0 (
    echo Generation completed successfully.
) else (
    echo Error during generation. Exit code: %ERRORLEVEL%
)

pause
