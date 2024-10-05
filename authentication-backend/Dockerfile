FROM golang:1.23 AS builder

WORKDIR /app
RUN mkdir /artifacts

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY authentication-backend authentication-backend

RUN CGO_ENABLED=0 GOOS=linux go build -o /artifacts/authentication-backend authentication-backend/main.go

FROM scratch 

COPY --from=builder /artifacts/authentication-backend authentication-backend

ENTRYPOINT [ "./authentication-backend" ]
CMD [ "serve" ]
EXPOSE 8090