# Chapter 10

Author's Reference Code: <https://github.com/awoodbeck/gnp/tree/master/ch10>

## Note

The author has the source code for the Caddy modules referenced as
`github.com/[username]/[module]`. I opted instead to nest the modules within the
chapter directory. In order for the modules to be used within `main.go`, I have
used the following terminal commands to add local references within `go.mod`.

```console
go mod edit -replace github.com/nicholas-fedor/Network-Programming-with-Go/Ch10/caddy-restrict-prefix=./caddy-restrict-prefix
go mod edit -replace github.com/nicholas-fedor/Network-Programming-with-Go/Ch10/caddy-toml-adapter=./caddy-toml-adapter
```
