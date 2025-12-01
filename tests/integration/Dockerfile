FROM golang:1.22-alpine

WORKDIR /app

# Install required packages
RUN apk add --no-cache make curl git

# Copy test files (built from tests/integration/ directory)
COPY . .

# Initialize dependencies
RUN go mod tidy && go mod download

# Set default environment variables
ENV INTERX_URL=http://3.123.154.245:11000
ENV TEST_ADDRESS=kira143q8vxpvuykt9pq50e6hng9s38vmy844n8k9wx
ENV VALIDATOR_ADDRESS=kira1vvcj9avffvyav83gmptdlzrprgvsrjxzh7f9sz
ENV DELEGATOR_ADDRESS=kira177lwmjyjds3cy7trers83r4pjn3dhv8zrqk9dl

# Default command runs all tests
CMD ["go", "test", "-v", "-count=1", "./..."]
