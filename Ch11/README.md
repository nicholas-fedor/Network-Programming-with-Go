# Ch11

## Prerequisite

Keys are needed for this chapter.

Ensure `GOROOT` environment variable is set.

```console
echo $GOROOT
```

Expected result:

```console
/usr/local/go
```

If not, then set `GOROOT` by using the following:

```console
GOROOT=$(go env GOROOT)
```

Then, generate the needed keys. Run the following in the terminal from the `Ch11` directory.

```console
go run $GOROOT/src/crypto/tls/generate_cert.go -host localhost -ecdsa-curve P256
```

This should output `cert.pem` and `key.pem` to the directory.

The keys submitted to this repo are only for testing purposes and not to be used
for production.

## Listing 11-15: Generating a certificate and private-key pair for the server and the client

Use the preceding code to generate certificate and key pairs for both the server
and the client.

```console
go run cert/generate.go -cert serverCert.pem -key serverKey.pem -host localhost
```

```console
go run cert/generate.go -cert clientCert.pem -key clientKey.pem -host localhost
```
