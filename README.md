# libp2p-facade

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://pkg.go.dev/github.com/amirylm/libp2p-facade?tab=doc)
![Github Actions](https://github.com/amirylm/libp2p-facade/actions/workflows/test.yml/badge.svg?branch=main)

Utilities plus a facade interface on top of 
[libp2p](https://github.com/libp2p/go-libp2p) host and major components:

- Streams were simplfied into a `Request` and `Handle` procedues
- Pubsub can be used with a simpler api to avoid topic management
- Config has a simple and extensible structure, inspired by 
[go-libp2p/config#Config](https://pkg.go.dev/github.com/libp2p/go-libp2p/config#Config)
- Metrics (prometheus)

## Install 

```shell
go get github.com/amirylm/libp2p-facade
```
