@echo off
REM Proxy Envoy en Windows usando Docker Desktop (el envoyServer del profesor es un binario de Linux).
REM Traduce gRPC-Web (HTTP/1.1, puerto 8080) a gRPC (HTTP/2) hacia el Servidor de Streaming (puerto 50051).
cd /d "%~dp0"
docker run --rm --name proxy-envoy -p 8080:8080 ^
  -v "%cd%\envoyConfigDocker.yaml:/etc/envoy/envoy.yaml" ^
  envoyproxy/envoy:v1.36.2 -c /etc/envoy/envoy.yaml --disable-hot-restart
