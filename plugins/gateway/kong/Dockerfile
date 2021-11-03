FROM kong/go-plugin-tool:2.0.4-alpine-latest AS builder
RUN mkdir -p /tmp/plugin/
COPY . /tmp/plugin/
RUN cd /tmp/plugin/ && \
   rm go.* && \
   go get github.com/Kong/go-pdk && \
   go mod init kong-go-plugin && \
   go get -d -v github.com/Kong/go-pluginserver && \
   go build github.com/Kong/go-pluginserver && \
   go build -buildmode plugin plugin.go

FROM scratch
WORKDIR /app

COPY --from=builder ["/tmp/plugin/plugin.so", "/plugin.so"]
