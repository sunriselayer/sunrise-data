# Sunrise Data

This software "sunrise-data" is a program to publish and retrieve the BLOB data from off chain storage like IPFS, Arweave and so on.

## Run sunrise chain node

## Run Local IPFS Node

sunrise-data uses IPFS protocol to upload metadata and shard to IPFS.

### 1. Download and extract ipfs node binary

You can download official prebuilt binaries from [IPFS kubo](https://dist.ipfs.tech#kubo) and Extract `kubo_v0.31.0_linux-amd64.tar.gz` after download.

```bash
wget https://dist.ipfs.tech/kubo/v0.29.0/kubo_v0.29.0_darwin-amd64.tar.gz
```

### 2. Install IPFS

macOS

```bash
mv ./kubo/ipfs /usr/local/bin/
```

Linux

```bash
wget https://dist.ipfs.tech/kubo/v0.31.0/kubo_v0.31.0_linux-amd64.tar.gz
tar -xvzf kubo_v0.31.0_linux-amd64.tar.gz
cd kubo
sudo ./install.sh
```

To check ipfs has been installed.

```bash
ipfs version
　ipfs version 0.31.0
```

### 3. Initialize IPFS

```bash
ipfs init --profile=lowpower
```

### 4. Run IPFS as daemon

```bash
ipfs daemon
```

### 5. Get IPFS node id

```bash
ipfs id
```

### 6. Add a remote IPFS node to the peer list

```bash
$ ipfs bootstrap add /ip4/13.114.102.20/tcp/4001/p2p/12D3KooWSBJ1warTMHy7bdaViev6udyWU8XBnz9QCYS8TSX9qadt
```

You can visit `http://localhost:8080/ipfs` to check runing IPFS RPC.

## Store project

This project need to store in directory stored "sunrise" codebase

```bash
ls
　sunrise
  sunrise-data
```

## Run API Service

- Prepare config.toml
  Copy config.default.toml to config.toml and replace your configurations.

  To connect to a local IPFS daemon, leave the `ipfs_api_url` field empty. 
　For a remote IPFS daemon, specify the HTTP URL along with the rpc port number, e.g. `http://1.2.3.4:5001`.

　`home_path` is your `sunrised` installed path.
　`publisher_account` is your publisher account name in your `sunrised` keyring.

  if you are not a validator, leave `proof_deputy_account` & `validator_address` empty.

- Run daemon

```sh
make dev
```

- Install daemon

```sh
make install
sunrise-data api # if you use api service for OP-Stack, etc.
sunrise-data rollkit # if you publish data from rollkit
sunrise-data validator # if you are a validator
```

## API Endpoint

### 1. POST `http://localhost:8000/api/publish`

Request JSON:

```protobuf
{
    "blob": "Base64 Encoded string",
    "data_shard_count": number,
    "parity_shard_count": number,
    "protocol": "ipfs" or "arweave"
}
```

Response JSON:

```protobuf
{
    metadata_uri: "metadata_uri"
}
```

### 2. GET `http://localhost:8000/api/shard-hashes?metadata_uri=[metadata_uri]&indices=1,2,3`

Response:

```protobuf
{
    shard_size: number,
    shard_uris:[
        "uri1",
        "uri2",
        ...
    ],
    shard_hashes:[
        "base64 encoded shard for index1",
        "base64 encoded shard for index2",
        "base64 encoded shard for index3",
        ...
    ]
}
```

### 3. GET `http://localhost:8000/api/get-blob?metadata_uri`

Response:

```protobuf
{
    blob:"base64 encoded blob",
}
```

### 4. Issue on API

In case that error occurs on API service, Endpoint returns HTTP 400 code and error msg.

## Testing Example for API Endpoint

### 1. Publsh

```protobuf
POST http://localhost:8000/api/publish
Request
{
    "blob": "MTIzNDU2Nzg5MDEyMzQ1Njc4OTA=",   // "12345678901234567890"
    "data_shard_count": 5,
    "parity_shard_count": 5,
    "protocol": "ipfs" // or "arweave"
}

Response
{
    metadata_uri: "ipfs://QmPXFt19HTkGjoZcbavLEgYYsuPm2xJR7hkhQxtRgPURMU"
}
```

### 2. Shard Hashes API

```protobuf
GET http://localhost:8000/api/shard-hashes?metadata_uri=ipfs://QmPdJ4GtFRvpkbsn47d1HbEioSYtSvgAYDkq5KsL5xUb1C&indices=1,2,3

{
    "shard_size":7,
    "shard_uris":[
        "ipfs://QmYbgKse7s4S1qSrz139zsPECYSu9HbHuz9TBy7ZDEKi54",
        "ipfs://QmWGhZL3maoUPbaYNauhq4BLL33xZdrf9Bi7xHUMFgtnV7",
        "ipfs://QmapUiNguJpqfuWdxtUJ1GPp5264yCLN5aMJqgWJvxdaEu",
        "ipfs://QmXLtGEkGVcRZukdaftW3M979SPWDaZidt6EpjkEk4SjCv",
        "ipfs://QmQzrZhSG3hAfwfJinMjiwC66MnJV6LxaVtLNKCjWaRdmj",
        "ipfs://Qmd2tYLCM7YecjoLA9ppJNPRgQnshVfdoiPwnux3WiyS2H"
    ],
    "shard_hashes":[
        "JvpetAD7cXIa6zMnWYOyOCfYD+g68xbHBVU5CEKz9OI=",
        "DKLinTzcoegAW/1rEIfBswH0ZXu6+W0V01PZb83Xmzg=",
        "A0+kyUqS08YRt34emU3OISrcCOWn3z7kCkBzftiKqog="
    ]
}
```

### 3. Get blob data from metadata_uri

```protobuf
GET http://localhost:8000/api/get-blob?metadata_uri=ipfs://QmPdJ4GtFRvpkbsn47d1HbEioSYtSvgAYDkq5KsL5xUb1C

{
    "blob": "MTIzNDU2Nzg5MDEyMzQ1Njc4OTA"
}
```

## Monitoring Service

- Search transactions
- Verify shard double hashes in published data
- Submit MsgChallengeForFraud
- Submit MsgSubmitProof
