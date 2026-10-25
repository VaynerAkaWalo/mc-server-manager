FROM --platform=$BUILDPLATFORM node:23.11.0-alpine AS frontend

WORKDIR /build

COPY web/package.json web/package-lock.json ./

RUN npm ci

COPY web .

RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine AS build

ARG TARGETARCH

WORKDIR /build

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o app cmd/main.go

FROM golang:1.24.2-alpine

WORKDIR /app

COPY --from=build build/app .

COPY --from=frontend /build/dist static

ENTRYPOINT ["./app"]
