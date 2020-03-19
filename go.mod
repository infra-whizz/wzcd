module github.com/infra-whizz/wzcd

go 1.13

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/infra-whizz/wzlib v0.0.0-20200318221301-86085037f399
	github.com/isbm/go-nanoconf v0.0.0-20200219130459-fc328232826c
	github.com/nats-io/jwt v0.3.2 // indirect
	github.com/nats-io/nats.go v1.9.1
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6 // indirect
)

replace github.com/infra-whizz/wzlib => /home/bo/work/golang/infra-whizz/wzlib
