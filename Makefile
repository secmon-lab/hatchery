MOCK_FILE=mock/mock_gen.go

all: mock

mock: $(MOCK_FILE)

$(MOCK_FILE): interfaces.go
	go run github.com/matryer/moq@latest -pkg mock -out $(MOCK_FILE) . Source Destination
