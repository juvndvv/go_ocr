# Etapa de construcción
FROM golang:latest as builder

# Instalar air
RUN go install github.com/air-verse/air@latest

# Instalar dependencias del sistema
RUN apt-get update && \
    apt-get install -y poppler-utils && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY ./config/air/.air.conf ./config/air/
COPY . .

# Etapa final
FROM golang:latest

# Copiar solo air desde la etapa builder
COPY --from=builder /go/bin/air /go/bin/air
COPY --from=builder /usr/bin/pdftotext /usr/bin/pdftotext

WORKDIR /app
COPY . .

EXPOSE 8080
CMD ["air"]