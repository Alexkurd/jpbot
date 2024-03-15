FROM golang:1.21-alpine as gobuild

COPY . /src
WORKDIR /src

RUN go build jpbot

FROM gcr.io/distroless/static
COPY --from=gobuild /src/*.yaml /src/jpbot.* /src/jpbot /
CMD ["./jpbot"]