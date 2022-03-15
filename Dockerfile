# Build the manager binary
FROM golang:1.16 as builder

ENTRYPOINT ["/manager"]
