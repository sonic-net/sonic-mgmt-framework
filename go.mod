module github.com/Azure/sonic-mgmt-framework

require (
	github.com/Azure/sonic-mgmt-common v0.0.0-00010101000000-000000000000
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.4
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/pkg/profile v1.4.0
	golang.org/x/crypto v0.17.0
	gopkg.in/go-playground/validator.v9 v9.31.0
)

replace github.com/Azure/sonic-mgmt-common => ../sonic-mgmt-common

go 1.13
