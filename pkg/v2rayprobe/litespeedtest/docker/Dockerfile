# sudo docker build   --network=host  -t lite:0.1 -f ./docker/Dockerfile ./
FROM golang:1.19 as wasm
ADD / /wasm/lite
WORKDIR /wasm/lite

FROM node:16 as gui
ADD / /gui/lite
WORKDIR /gui/lite


FROM golang:1.19 as build
ADD / /go/lite
WORKDIR /go/lite

# china only
# RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN make dockerbin

FROM gcr.io/distroless/static-debian11
# FROM gcr.lank8s.cn/distroless/static-debian11
COPY --from=build /go/lite/bin/lite-dockerbin /
EXPOSE 10888
CMD [ "/lite-dockerbin" ]