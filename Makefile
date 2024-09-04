MOCK_HATCHERY=pkg/mock/hatcnery_gen.go
MOCK_PKG=pkg/mock/pkg_gen.go

all: mock

mock: $(MOCK_HATCHERY) $(MOCK_PKG)

$(MOCK_HATCHERY): interfaces.go
	go run github.com/matryer/moq@latest -pkg mock -out $(MOCK_HATCHERY) . Source Destination

$(MOCK_PKG): interfaces.go ./pkg/interfaces/*
	go run github.com/matryer/moq@latest -pkg mock -out $(MOCK_PKG) ./pkg/interfaces HTTPClient SQS S3
