FROM scratch
COPY load-simulator /
ENTRYPOINT ["/load-simulator"]
