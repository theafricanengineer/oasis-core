module github.com/oasislabs/oasis-core/go

replace (
	github.com/tendermint/tendermint => github.com/oasislabs/tendermint v0.32.0-dev2.0.20200430124658-ebbd5250a562
	golang.org/x/crypto/curve25519 => github.com/oasislabs/ed25519/extra/x25519 v0.0.0-20191022155220-a426dcc8ad5f
	golang.org/x/crypto/ed25519 => github.com/oasislabs/ed25519 v0.0.0-20191109133925-b197a691e30d
)

require (
	github.com/RoaringBitmap/roaring v0.4.18 // indirect
	github.com/blevesearch/bleve v0.8.0
	github.com/blevesearch/blevex v0.0.0-20180227211930-4b158bb555a3 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.2 // indirect
	github.com/blevesearch/segment v0.0.0-20160915185041-762005e7a34f // indirect
	github.com/cenkalti/backoff/v4 v4.0.0
	github.com/couchbase/vellum v0.0.0-20190610201045-ec7b775d247f // indirect
	github.com/cznic/b v0.0.0-20181122101859-a26611c4d92d // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/eapache/channels v1.1.0
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fxamacker/cbor/v2 v2.2.0
	github.com/glycerine/goconvey v0.0.0-20190410193231-58a59202ab31 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/golang/protobuf v1.4.0
	github.com/golang/snappy v0.0.1
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/hpcloud/tail v1.0.0
	github.com/libp2p/go-libp2p v0.1.1
	github.com/libp2p/go-libp2p-core v0.0.3
	github.com/libp2p/go-msgio v0.0.3 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/multiformats/go-multihash v0.0.6 // indirect
	github.com/oasislabs/deoxysii v0.0.0-20190807103041-6159f99c2236
	github.com/oasislabs/ed25519 v0.0.0-20191122104632-9d9ffc15f526
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/common v0.9.1
	github.com/prometheus/procfs v0.0.11
	github.com/remyoudompheng/bigfft v0.0.0-20190512091148-babf20351dd7 // indirect
	github.com/seccomp/libseccomp-golang v0.9.1
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/steveyen/gtreap v0.0.0-20150807155958-0abe01ef9be2 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.8
	github.com/tendermint/tm-db v0.5.1
	github.com/thepudds/fzgo v0.2.2
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/whyrusleeping/go-logging v0.0.0-20170515211332-0457bb6b88fc
	github.com/zondax/ledger-oasis-go v0.3.0
	gitlab.com/yawning/dynlib.git v0.0.0-20190911075527-1e6ab3739fd8
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.29.1
	google.golang.org/grpc/security/advancedtls v0.0.0-20200406221258-98e4c7ad3eef
	google.golang.org/protobuf v1.21.0
	gopkg.in/ini.v1 v1.55.0 // indirect
)

go 1.13
