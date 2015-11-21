FROM gliderlabs/alpine:latest

COPY routed/routed /
RUN chmod +x /routed

ENTRYPOINT ["./routed"]