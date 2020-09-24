# tty-proxy

## Building the gobindata.go file

All files under ~assets/*~ are packed to the gobindata.go file which will be statically compiled
within the final binary.

```bash
	go get github.com/go-bindata/go-bindata/...
	go-bindata --prefix static -o gobindata.go static/*
```
