module github.com/pandodao/botastic

go 1.20

replace gorm.io/gen v0.3.22-0.20230213112745-6dc1af4c110b => github.com/pandodao/gen v0.0.0-20230215020331-916e2ec8363e

replace github.com/milvus-io/milvus-sdk-go/v2 v2.2.0 => github.com/pandodao/milvus-sdk-go/v2 v2.0.0-20230327110027-9f97cf0e4642

require (
	github.com/PuerkitoBio/goquery v1.8.1
	github.com/drone/signal v1.0.0
	github.com/ethereum/go-ethereum v1.11.2
	github.com/fox-one/pkg/logger v0.0.2
	github.com/go-chi/chi v1.5.4
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/lib/pq v1.10.7
	github.com/milvus-io/milvus-sdk-go/v2 v2.2.0
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/pandodao/tokenizer-go v0.0.1
	github.com/pressly/goose/v3 v3.9.0
	github.com/rs/cors v1.8.3
	github.com/sashabaranov/go-openai v1.5.7
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/tidwall/gjson v1.14.4
	golang.org/x/sync v0.1.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/postgres v1.4.5
	gorm.io/gen v0.3.22-0.20230213112745-6dc1af4c110b
	gorm.io/gorm v1.24.5
	gorm.io/plugin/dbresolver v1.4.1
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.2 // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/dop251/goja v0.0.0-20230304130813-e2f543bf4b4c // indirect
	github.com/dop251/goja_nodejs v0.0.0-20230226152057-060fa99b809f // indirect
	github.com/fkgi/abnf v1.0.0 // indirect
	github.com/fox-one/msgpack v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.18.0 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/pprof v0.0.0-20230309165930-d61513b1440d // indirect
	github.com/gorilla/schema v1.2.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/milvus-io/milvus-proto/go-api v0.0.0-20230303054144-16f081962572 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/relvacode/iso8601 v1.3.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/twitchtv/twirp v8.1.2+incompatible // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221227171554-f9683d7f8bef // indirect
	google.golang.org/grpc v1.52.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/driver/mysql v1.4.6 // indirect
)

require (
	github.com/fox-one/mixin-sdk-go v1.7.3
	github.com/fox-one/passport-go v1.0.4
	github.com/fox-one/pkg/httputil v0.0.1
	github.com/fox-one/pkg/uuid v0.0.1 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.8.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.13.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.15.0
	github.com/ugorji/go/codec v1.2.7 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/datatypes v1.1.1-0.20230130040222-c43177d3cf8c // indirect
	gorm.io/hints v1.1.1 // indirect
)
