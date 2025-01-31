# Cross-Compile for Different Platforms
If you want to build for multiple operating systems, use GOOS and GOARCH:

## Compile for Windows from Linux/macOS:

```sh
GOOS=windows GOARCH=amd64 go build -o PortHunter.exe main.go
```

## Compile for Linux from Windows/macOS:

```sh
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o PortHunter main.go
```

## Compile for macOS from Linux/Windows:

```sh
GOOS=darwin GOARCH=amd64 go build -o PortHunter main.go
```


## Strip Debug Information (Smaller Size)
To reduce the binary size, use:

```sh
go build -ldflags "-s -w" -o PortHunter main.go
```
-s: Removes the symbol table.
-w: Removes debug information.
