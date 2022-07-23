
FROM golang:alpine as builder

# Setup app folder
RUN mkdir /app
ADD . /app
WORKDIR /app

# Build the app
RUN go mod download && \
    go build -o crazed-nft-fans ./main . && \
    chmod +x ./crazed-nft-fans

# Move to small scratch image
FROM scratch
COPY --from=builder /app/crazed-nft-fans /app/crazed-nft-fans
EXPOSE 6060

ENTRYPOINT ["/app/crazed-nft-fans"]
