# Use an official Golang runtime as a parent image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the local files to the container's working directory
COPY . .

# Build the Go app
RUN go build -o main .

# Run the executable
CMD ["./main"]
