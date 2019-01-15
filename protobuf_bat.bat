@echo off
cd d:\sheshe
protoc --go_out=././src/dq/ protobuf/msg.proto
protoc --csharp_out=D:\unity3d\dotaNet\Assets\Script\ protobuf/msg.proto
pause