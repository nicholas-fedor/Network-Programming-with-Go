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
