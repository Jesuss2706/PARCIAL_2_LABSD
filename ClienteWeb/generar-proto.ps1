# Regenera src\serviciosAudio\audio_pb.js y audio_grpc_web_pb.js a partir de proto\audio.proto.
# Solo es necesario si cambia el .proto (los archivos generados ya vienen incluidos).
#
# Requiere en el PATH (o en esta carpeta):
#   protoc.exe                 https://github.com/protocolbuffers/protobuf/releases        (protoc-XX-win64.zip)
#   protoc-gen-js.exe          https://github.com/protocolbuffers/protobuf-javascript/releases (protobuf-javascript-3.21.4-win64.zip)
#   protoc-gen-grpc-web.exe    https://github.com/grpc/grpc-web/releases                   (protoc-gen-grpc-web-1.5.0-windows-x86_64.exe, renombrar)
Set-Location $PSScriptRoot
protoc -I=proto audio.proto `
  --js_out=import_style=commonjs:src/serviciosAudio `
  --grpc-web_out=import_style=commonjs,mode=grpcwebtext:src/serviciosAudio
npx webpack
