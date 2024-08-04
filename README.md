# Sunrise Data

This software "sunrise-data" is a program to publish and retrieve the BLOB data from off chain storage like IPFS, Arweave and so on.

## Store project
This project need to store in directory stored "cau-sunrise" codebase
```bash
> ls
  cau-sunrise
  cau-sunrise-data
```

## Run API Service
```sh
go mod tidy
go run .
```

## API Endpoint

### 1. POST http://localhost:8080/api/publish

Request JSON:
```protobuf
{
    "blob": "Base64 Encoded string",
    "shard_count_half": number
}
```
Response JSON:
```protobuf
{
    metadata_uri: "metadata_uri"
}
```

### 2. GET http://localhost:8080/api/uploaded_data?metadata_uri=[metadata_uri]

Response:
```protobuf
{
    shard_size: number,
    shard_count: number,
    shard_uris:[
        "uri1",
        "uri2",
        ...
    ]
}
```

