FROM alpine:latest

COPY build/k8s-secret-creator /
COPY test/secret.yaml /

RUN chmod +x /k8s-secret-creator

ENTRYPOINT ["/k8s-secret-creator"]
