module example/sawtooth-plugin

go 1.13

require (
	github.com/golang/protobuf v1.4.3
	github.com/hyperledger/sawtooth-sdk-go v0.1.0
	github.com/jessevdk/go-flags v1.4.0
	//github.com/hyperledger/fabric-lib-go v1.0.0 // indirect
	github.com/meshplus/bitxhub-model v1.0.0-rc4.0.20200608065824-2fbc63639e92
	github.com/meshplus/pier v0.0.0-00010101000000-000000000000
	github.com/pebbe/zmq4 v1.2.7 // indirect
	gopkg.in/yaml.v2 v2.4.0
	//github.com/btcsuite/btcd/btcec v0.21.0-beta
)

replace github.com/golang/protobuf => github.com/golang/protobuf v1.3.2

replace google.golang.org/grpc => google.golang.org/grpc v1.27.1

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.7

replace golang.org/x/net => golang.org/x/net v0.0.0-20200202094626-16171245cfb2

replace github.com/meshplus/pier => ../pier

replace golang.org/x/text => golang.org/x/text v0.3.0

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200509044756-6aff5f38e54f

replace github.com/spf13/afero => github.com/spf13/afero v1.1.2

replace github.com/spf13/pflag => github.com/spf13/pflag v1.0.5

replace github.com/pelletier/go-toml => github.com/pelletier/go-toml v1.2.0

replace github.com/spf13/jwalterweatherman => github.com/spf13/jwalterweatherman v1.0.0

replace github.com/btcsuite/btcd/btcec => github.com/btcsuite/btcd/btcec v0.20.1-beta

replace encoding => /usr/local/go/pkg/linux_amd64/encoding

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200218151345-dad8c97a84f5
