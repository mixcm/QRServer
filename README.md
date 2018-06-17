# QRServer
Generate and cache QR codes in a Go way.

## Installation

```bash
go get github.com/mixcm/QRServer
```

## Run

```bash
QRServer -port=[Port to listen] -node=[Name of current node]
```

Example:

```bash
QRServer -port=:23184 -node=China.Beijing.01
QRServer -port=127.0.0.1:28080 -node=US.LosAngeles.02
```