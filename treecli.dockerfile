FROM obraun/vss-protoactor-jenkins as builder
COPY . /app
WORKDIR /app
RUN go build -o treecli/treecli treecli/main.go

FROM iron/go
COPY --from=builder /app/treecli/treecli /app/treecli
EXPOSE 8090
ENTRYPOINT [ "/app/treecli" ]
