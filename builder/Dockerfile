FROM gcr.io/cloud-builders/go AS build-env

ADD ./ /go/src/

ENV GOBIN=/go/bin
RUN go get -d -v ./...
RUN go install /go/src/main.go

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/main /usr/bin/main