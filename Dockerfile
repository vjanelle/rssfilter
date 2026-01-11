# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/rssproxy ./

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/rssproxy /rssproxy

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/rssproxy"]
