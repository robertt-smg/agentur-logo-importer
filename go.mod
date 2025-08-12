module agentur-logo-importer

go 1.23.0

toolchain go1.23.6

replace github.com/robertt-smg/agentur-logo-importer => ./

require (
	github.com/Azure/go-ntlmssp v0.0.0-20180810175552-4a21cbd618b4
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/h2non/filetype v1.0.8
	github.com/joho/godotenv v1.5.1
	github.com/kpango/glg v1.6.4
	github.com/mdouchement/hdr v0.2.1
	github.com/pkg/sftp v1.10.0
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/image v0.0.0-20190703141733-d6a02ce849c9
	gopkg.in/ini.v1 v1.42.0
)

require (
	github.com/goccy/go-json v0.7.4 // indirect
	github.com/kpango/fastime v1.0.17 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.0.1 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	gonum.org/v1/gonum v0.0.0-20181125185008-b630de2f2264 // indirect
	google.golang.org/appengine v1.6.1 // indirect
)

replace github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v1.0.0

replace github.com/nats-io/go-nats => github.com/nats-io/nats.go v1.8.1

replace github.com/circonus-labs/circonusllhist => github.com/openhistogram/circonusllhist v0.2.1

replace github.com/pressly/chi => github.com/go-chi/chi v1.5.4

replace github.com/codegangsta/cli => github.com/urfave/cli v1.22.5

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5

replace github.com/lyft/protoc-gen-validate => github.com/envoyproxy/protoc-gen-validate v0.6.1

replace github.com/apache/thrift/lib/go/thrift => github.com/apache/thrift v0.13.0
