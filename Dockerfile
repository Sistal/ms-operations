# Etapa de construcción
FROM golang:1.24-alpine AS builder

# Establecer directorio de trabajo
WORKDIR /app

# Instalar dependencias del sistema necesarias para compilar (si las hay)
# RUN apk add --no-cache git

# Copiar archivos de definición de módulos
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
# CGO_ENABLED=0 deshabilita cgo para asegurar un binario estático
# GOOS=linux asegura que el binario sea para Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o ms-operations cmd/api/main.go

# Etapa final
FROM alpine:latest

# Instalar certificados CA para llamadas HTTPS y bash/curl para depuración si es necesario
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/ms-operations .

# Copiar migraciones si son necesarias en tiempo de ejecución (opcional, ajusta según necesidad)
# COPY --from=builder /app/migrations ./migrations

# Exponer el puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./ms-operations"]
