# Kool
The tool that does more with less.

# Install

Plain `go install`:
```sh
go install github.com/xhd2015/kool@latest
```

`curl` from github:
```sh
curl -fsSL https://github.com/xhd2015/kool/raw/master/install.sh | bash
```

# Usage

Kool commands are splited into the following categories:
- string
- json
- git
- go
- net
- others

## string
```sh
# uniq lines
pbpaste | kool lines uniq
```

## json
```sh
pbpaste | kool compress
pbpaste | kool pretty 
```

## go

Debug helper:
```sh
# automatically spawn a dlv process
kool go run --debug ./
```

Version helper:
```sh
# run with specific version
kool with-go1.22 go run ./
```

## git

```sh
# tag HEAD with next version, and push
kool git tag-next --push

# show HEAD's version
kool git show-tag
```

## net
```sh
# kill the porcess that listens on 8080
kool kill-port 8080
```