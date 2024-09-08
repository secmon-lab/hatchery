MOCK_PKG=pkg/mock/pkg_gen.go

all: mock

mock: $(MOCK_PKG)

$(MOCK_PKG): interfaces.go ./pkg/interfaces/*
	go run github.com/matryer/moq@latest -pkg mock -out $(MOCK_PKG) ./pkg/interfaces HTTPClient SQS S3
