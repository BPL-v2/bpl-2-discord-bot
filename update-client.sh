#!/bin/bash
# Get spec from local bpl api server
URL="http://localhost:8000/api/swagger/doc.json"

# Download the OpenAPI specification and pipe it to the Swagger converter API to turn it into a 3.0 specification
curl $URL | curl -X POST -H "Content-Type: application/json" -d @- https://converter.swagger.io/api/convert -o tools/swagger.json

# Run the go generate command
go generate tools/tools.go

# Remove the spec file
rm -rf tools/swagger.json
