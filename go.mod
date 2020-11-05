module github.com/datachainlab/cross

go 1.15

require (
	github.com/bluele/interchain-simple-packet v0.0.0-20200921011237-118864bc041e
	github.com/cosmos/cosmos-sdk v0.40.0-rc2.0.20201104140222-71166c894920
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.15.2
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.34.0-rc5
	github.com/tendermint/tm-db v0.6.2
	google.golang.org/genproto v0.0.0-20201019141844-1ed22bb0c154 // indirect
	google.golang.org/grpc v1.33.1
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
