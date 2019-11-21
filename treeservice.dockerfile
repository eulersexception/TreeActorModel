FROM obraun/vss-protoactor-jenkins as builder
COPY . /app
WORKDIR /app
RUN go build -o treeservice/treeservice treeservice/main.go

FROM iron/go
COPY --from=builder /app/treeservice/treeservice /app/treeservice
EXPOSE 8091
ENTRYPOINT ["/app/treeservice"]
