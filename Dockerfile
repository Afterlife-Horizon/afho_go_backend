FROM debian:latest


# Get the dependencies
RUN set -x \
  && apt-get update \ 
  && apt-get install ffmpeg -y

COPY --from=golang:1.21-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"


WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /afho_backend

EXPOSE 4000
EXPOSE 80
CMD ["/afho_backend"]
