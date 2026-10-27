@echo off
REM Servidor de reacciones (Spring Boot, WebSocket) en el puerto 8085. Requiere JDK 17 o superior.
cd /d "%~dp0"
call mvnw.cmd spring-boot:run
