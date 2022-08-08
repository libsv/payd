module github.com/libsv/payd

go 1.18

require (
	github.com/boombuler/barcode v1.0.1
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmoiron/sqlx v1.3.5
	github.com/labstack/echo/v4 v4.7.2
	github.com/libsv/go-bk v0.1.6
	github.com/libsv/go-bt/v2 v2.1.0-beta.3
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-sqlite3 v1.14.10
	github.com/pkg/errors v0.9.1
	github.com/speps/go-hashids v2.0.0+incompatible
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.8.0
	github.com/swaggo/echo-swagger v1.3.3
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	github.com/swaggo/swag v1.8.4
	github.com/theflyingcodr/govalidator v0.1.3
	github.com/theflyingcodr/lathos v0.0.6
	github.com/tonicpow/go-minercraft v0.8.0
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	gopkg.in/guregu/null.v3 v3.5.0
)

require (
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/libsv/go-bc v0.1.10
	github.com/rs/zerolog v1.27.0
	github.com/theflyingcodr/sockets v0.0.12-beta
)

require (
	github.com/InVisionApp/go-logger v1.0.1 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
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
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/subosito/gotenv v1.3.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/net v0.0.0-20220728030405-41545e8bf201 // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	golang.org/x/tools v0.1.10 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/InVisionApp/go-health/v2 v2.1.2
	github.com/labstack/echo-contrib v0.13.0
	github.com/libsv/go-dpp v0.1.11
	github.com/libsv/go-spvchannels v0.0.2
)

replace github.com/golang-migrate/migrate/v4 => github.com/theflyingcodr/migrate/v4 v4.15.1-0.20210927160112-79da889ca18e
