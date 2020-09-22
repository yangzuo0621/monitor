FROM golang:1.13.15-buster AS build-stage
ENV CURRENT_MODULE github.com/yangzuo0621/azure-devops-cmd/monitor
COPY . /go/src/$CURRENT_MODULE
RUN cd /go/src/$CURRENT_MODULE/cmd/monitor && go build -o /tmp/monitor .

FROM gcr.io/distroless/base-debian10:debug
COPY --from=build-stage /tmp/monitor /monitor

ENTRYPOINT [ "/monitor" ]