# Build the backend
FROM golang:alpine as build-backend

COPY . /app
WORKDIR /app

RUN \
  apk add git gcc && \
  go build -installsuffix cgo -ldflags '-extldflags "-static"' -o app


# Build the frontend
FROM node:alpine as build-frontend

COPY --from=build-backend /app /app
WORKDIR /app/frontend
RUN yarn && yarn build

# The actual docker file
FROM alpine

RUN mkdir -p /app/frontend

COPY --from=build=frontend /app/app /app/app
COPY --from=build=frontend /app/frontend/build /app/frontend/build

WORKDIR /app

CMD ["./app"]
