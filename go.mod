module github.com/datachainlab/cross

go 1.13

require (
	github.com/cosmos/cosmos-sdk v0.34.4-0.20200305135144-04a1f6f8ca17
	github.com/gorilla/mux v1.7.4
	github.com/otiai10/copy v1.0.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.1
	github.com/tendermint/tm-db v0.4.1
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
