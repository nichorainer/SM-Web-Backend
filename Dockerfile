# Stage build
FROM golang:1.21 AS build
WORKDIR /app
COPY . .
RUN go build -o server ./cmd

# Stage run
FROM gcr.io/distroless/base-debian11
COPY --from=build /app/server /server
CMD ["/server"]