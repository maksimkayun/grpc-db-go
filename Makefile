include .env

LOCAL_BIN:=$(CURDIR)/bin
PROTOC_VERSION=24.4
PROTOC_ZIP=protoc-$(PROTOC_VERSION)-win64.zip

LOCAL_MIGRATION_DIR=$(MIGRATION_DIR)
LOCAL_MIGRATION_DSN="host=localhost port=$(PG_PORT) dbname=$(PG_DATABASE_NAME) user=$(PG_USER) password=$(PG_PASSWORD) sslmode=disable"

install-deps:
	if not exist "$(LOCAL_BIN)" mkdir "$(LOCAL_BIN)"
	powershell -Command "Invoke-WebRequest -Uri https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP) -OutFile $(LOCAL_BIN)/$(PROTOC_ZIP)"
	powershell -Command "Expand-Archive -Path '$(LOCAL_BIN)/$(PROTOC_ZIP)' -DestinationPath '$(LOCAL_BIN)/protoc-temp' -Force"
	copy "$(LOCAL_BIN)\protoc-temp\bin\protoc.exe" "$(LOCAL_BIN)\protoc.exe"
	if not exist "include" mkdir "include"
	xcopy /E /I "$(LOCAL_BIN)\protoc-temp\include" "include"
	rd /s /q "$(LOCAL_BIN)\protoc-temp"
	del "$(LOCAL_BIN)\$(PROTOC_ZIP)"
	go env -w GOBIN=$(LOCAL_BIN)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	go env -w GOBIN=$(LOCAL_BIN)
	go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	go env -w GOBIN=$(LOCAL_BIN)
	go install github.com/pressly/goose/v3/cmd/goose@v3.14.0

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go get -u github.com/pressly/goose/v3/cmd/goose

generate:
	if not exist pkg mkdir pkg
	if not exist pkg\note_v1 mkdir pkg\note_v1
	make generate-note-api

generate-note-api:
	$(LOCAL_BIN)\protoc.exe \
		--proto_path=api/note_v1 \
		--proto_path=include \
		--go_out=pkg/note_v1 \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg/note_v1 \
		--go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-go=$(LOCAL_BIN)\protoc-gen-go.exe \
		--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)\protoc-gen-go-grpc.exe \
		api/note_v1/note.proto

build:
	go env -w GOOS=windows && go env -w GOARCH=amd64 && go build -o service_windows cmd/grpc_server/main.go

start-server:
	go run cmd/grpc_server/main.go

start-client:
	go run cmd/grpc_client/main.go

local-migration-status:
	$(LOCAL_BIN)\goose -dir $(LOCAL_MIGRATION_DIR) postgres $(LOCAL_MIGRATION_DSN) status -v

local-migration-up:
	$(LOCAL_BIN)\goose -dir $(LOCAL_MIGRATION_DIR) postgres $(LOCAL_MIGRATION_DSN) up -v

local-migration-down:
	$(LOCAL_BIN)\goose -dir $(LOCAL_MIGRATION_DIR) postgres $(LOCAL_MIGRATION_DSN) down -v