# Goose

Goose (Go Online Storage Engine) is a simple static file server implemented in Go over MongoDB's GridFS spec.

## Quick usage examples

### Bucket creation


### File upload

In this example, we use cURL to upload a PDF file: 

1. The ID of the bucket we're uploading it to is *546e1759494d911a70000001*

2. The path under the bucket where the file will be accesible is */uploads/Book.pdf*

```
curl --tr-encoding -X POST -v -# -o output -T Book.pdf -H "Content-Type: application/pdf" \
  http://api.goose.loc:3000/buckets/546e1759494d911a70000001/objects?name=/uploads/Book.pdf
```

## Roadmap

- Whitelist for selective unmarshaling of JSON in structs
- Filter by bucket when serving a file
- Error and debug logging
- Data validation for POST / PUT
- JWT auth
- API Client for integration in other Go services
- Dockerize the app for easier deployment
- Web based client for uploading and browsing buckets
