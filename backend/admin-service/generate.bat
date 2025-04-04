@echo off

:: Ensure GOPATH\bin is in PATH for protoc plugins
set PATH=%PATH%;%GOPATH%\bin

:: Create proto directory if it doesn't exist
if not exist proto mkdir proto

:: Generate the protobuf and gRPC code
echo Generating protobuf and gRPC code...
protoc --proto_path=. ^
       --go_out=. ^
       --go_opt=paths=source_relative ^
       --go-grpc_out=. ^
       --go-grpc_opt=paths=source_relative ^
       proto\admin.proto

if %ERRORLEVEL% EQU 0 (
    echo Generation completed successfully.
) else (
    echo Error during generation. Exit code: %ERRORLEVEL%
)

pause