# Build the backend
FROM golang:alpine as build-backend

RUN mkdir -p /go/src/github.com/mjarkk/machinery-dashboard
WORKDIR /go/src/github.com/mjarkk/machinery-dashboard

COPY . /go/src/github.com/mjarkk/machinery-dashboard
RUN \
  apk add git gcc libc-dev && \
  go get && \
  go build -installsuffix cgo -ldflags '-extldflags "-static"' -o app


# Build the frontend
FROM node:alpine as build-frontend

COPY --from=build-backend /go/src/github.com/mjarkk/machinery-dashboard /app
WORKDIR /app/frontend
RUN yarn && yarn build


# The actual docker file
FROM alpine

WORKDIR /
RUN mkdir -p /frontend/build

COPY --from=build-frontend /app/app /app
COPY --from=build-frontend /app/frontend/build /frontend/build

CMD ["./app"]
