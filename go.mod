module github.com/libsv/payd

go 1.17

require (
	github.com/boombuler/barcode v1.0.1
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmoiron/sqlx v1.3.4
	github.com/labstack/echo/v4 v4.7.0
	github.com/libsv/go-bk v0.1.6
	github.com/libsv/go-bt/v2 v2.1.0-beta.2
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-sqlite3 v1.14.10
	github.com/pkg/errors v0.9.1
	github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/echo-swagger v1.3.0
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2 // indirect
	github.com/swaggo/swag v1.8.0
	github.com/theflyingcodr/govalidator v0.1.3
	github.com/theflyingcodr/lathos v0.0.6
	github.com/tonicpow/go-minercraft v0.4.0
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/guregu/null.v3 v3.5.0
)

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/libsv/go-bc v0.1.8
	github.com/rs/zerolog v1.26.1
	github.com/theflyingcodr/sockets v0.0.11-beta.0.20220225103542-c6eecb16f586
)

require (
	github.com/InVisionApp/go-logger v1.0.1 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitcoinsv/bsvd v0.0.0-20190609155523-4c29707f7173 // indirect
	github.com/bitcoinsv/bsvutil v0.0.0-20181216182056-1d77cf353ea9 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gojektech/heimdall/v6 v6.1.0 // indirect
	github.com/gojektech/valkyrie v0.0.0-20190210220504-8f62c1e7ba45 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/libsv/go-bt v1.0.4 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	golang.org/x/tools v0.1.9 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

require (
	github.com/InVisionApp/go-health/v2 v2.1.2
	github.com/libsv/go-p4 v0.0.8
	github.com/libsv/go-spvchannels v0.0.1
)

replace github.com/golang-migrate/migrate/v4 => github.com/theflyingcodr/migrate/v4 v4.15.1-0.20210927160112-79da889ca18e
