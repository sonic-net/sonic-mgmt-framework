module github.com/Azure/sonic-mgmt-framework

require (
	github.com/Azure/sonic-mgmt-common v0.0.0-20261106021544-c49a9496f49b
	github.com/go-playground/validator/v10 v10.30.6
	github.com/golang/glog v1.2.5
	github.com/gorilla/mux v1.7.4
	github.com/pkg/profile v1.7.0
	golang.org/x/crypto v0.59.0
)

require (
	github.com/Workiva/go-datastructures v1.1.8 // indirect
	github.com/antchfx/jsonquery v1.1.4 // indirect
	github.com/antchfx/xmlquery v1.3.1 // indirect
	github.com/antchfx/xpath v1.3.6 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/godbus/dbus/v5 v5.2.4 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/maruel/natural v1.3.0 // indirect
	github.com/openconfig/gnmi v0.0.0-20200617225440-d2b4e6a45802 // indirect
	github.com/openconfig/goyang v0.0.0-20200309174518-a00bece872fc // indirect
	github.com/openconfig/ygot v0.7.1 // indirect
	github.com/philopon/go-toposort v0.0.0-20170620085441-9be86dbd762f // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go4.org/intern v0.0.0-20211027215823-ae77deb06f29 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20230525183740-e7c30c78aeb2 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260817212433-ac3dfec99bb1 // indirect
	google.golang.org/grpc v1.85.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	inet.af/netaddr v0.0.0-20230525184311-b8eac61e914a // indirect
)

go 1.26.0

// Pin glog to the version that patches/glog.patch is written for.
replace github.com/golang/glog => github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
