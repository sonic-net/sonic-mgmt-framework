module github.com/Azure/sonic-mgmt-framework

require (
	github.com/Azure/sonic-mgmt-common v0.0.0-00010101000000-000000000000
	github.com/golang/glog v1.2.5
	github.com/gorilla/mux v1.7.4
	github.com/pkg/profile v1.7.0
	golang.org/x/crypto v0.50.0
	gopkg.in/go-playground/validator.v9 v9.31.0
)

require (
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/antchfx/jsonquery v1.1.4 // indirect
	github.com/antchfx/xmlquery v1.3.1 // indirect
	github.com/antchfx/xpath v1.1.10 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/maruel/natural v1.1.1 // indirect
	github.com/openconfig/gnmi v0.0.0-20200617225440-d2b4e6a45802 // indirect
	github.com/openconfig/goyang v0.0.0-20200309174518-a00bece872fc // indirect
	github.com/openconfig/ygot v0.7.1 // indirect
	github.com/philopon/go-toposort v0.0.0-20170620085441-9be86dbd762f // indirect
	github.com/redis/go-redis/v9 v9.6.1 // indirect
	go4.org/intern v0.0.0-20211027215823-ae77deb06f29 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20230525183740-e7c30c78aeb2 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/grpc v1.82.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	inet.af/netaddr v0.0.0-20230525184311-b8eac61e914a // indirect
)

replace github.com/Azure/sonic-mgmt-common => ../sonic-mgmt-common

go 1.25.9

// Pin glog to the version that patches/glog.patch is written for.
replace github.com/golang/glog => github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
