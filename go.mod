module github.com/libsv/payd

go 1.16

require (
	github.com/bitcoinschema/go-bitcoin v0.3.17 // indirect
	github.com/boombuler/barcode v1.0.1
	github.com/coreos/bbolt v1.3.2 // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/jmoiron/sqlx v1.3.1
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/lib/pq v1.9.0 // indirect
	github.com/libsv/go-bc v0.1.4-0.20210907115037-52f4ad4cf321
	github.com/libsv/go-bk v0.1.4
	github.com/libsv/go-bt/v2 v2.0.0-beta.5.0.20210915144027-6e0ed1a6509f
	github.com/matryer/is v1.4.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.3 // indirect
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/theflyingcodr/govalidator v0.0.1
	github.com/theflyingcodr/lathos v0.0.3
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/tonicpow/go-minercraft v0.3.0
	github.com/tonicpow/go-paymail v0.1.6
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.2 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/tools v0.1.5 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/guregu/null.v3 v3.5.0
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/libsv/go-bt/v2 => ../go-bt
