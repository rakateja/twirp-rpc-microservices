gen:
	protoc --twirp_out=. --go_out=. proto/service.proto

swagger:
	twirp-swagger-gen -in proto/service.proto -out swaggerui/swagger.json -host localhost:9001

