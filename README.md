
# Build

go build main.go io.go consts.go shelf.go order.go configs.go

# test

go test
go test -race
go test -bench=.


## installation of shit

- (go 1.14)
- sudo snap install go --classic
- go get "github.com/orcaman/concurrent-map"
- go build main.go
