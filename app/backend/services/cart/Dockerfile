FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/cart ./services/cart/cmd/cart

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/cart /app/cart
EXPOSE 8086
ENTRYPOINT ["/app/cart"]
