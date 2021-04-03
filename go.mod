module gopkg.in/nullstone-io/nullstone.v0

go 1.16

require (
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gopkg.in/nullstone-io/go-api-client.v0 v0.0.0-00010101000000-000000000000
)

replace gopkg.in/nullstone-io/go-api-client.v0 => ../go-api-client
