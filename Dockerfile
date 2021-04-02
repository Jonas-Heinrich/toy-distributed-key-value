# Builder image
FROM golang:1.16.3 as builder

RUN mkdir /work
WORKDIR /work
COPY . .
RUN make build-kv

# Final image
FROM gcr.io/distroless/base

# Run without privileges
USER nobody:nobody

COPY --from=builder /work/build/toy-distributed-kv /go/bin/toy-distributed-kv 
ENTRYPOINT ["/go/bin/toy-distributed-kv"]
