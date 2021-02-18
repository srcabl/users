module github.com/srcabl/users

go 1.15

replace github.com/srcabl/services => /home/kero/automata/srcabl/services

replace github.com/srcabl/protos => /home/kero/automata/srcabl/protos

require (
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/srcabl/protos v0.1.0
	github.com/srcabl/services v0.1.1
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)
