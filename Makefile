all: bins

clean:
	@printf cleaning...
	@GO111MODULE=on go clean
	@echo done!

bins:
	@printf building...
	@GO111MODULE=on go build github.com/freerware/obscurer
	@echo done!

test: bins
	@echo testing...
	@GO111MODULE=on go test -v -race -covermode=atomic -coverprofile=obscurer.coverprofile github.com/freerware/obscurer

mocks:
	@mockgen -source=store.go -destination=./internal/mock/store.go -package=mock -mock_names=Store=Store

benchmark: bins
	@GO111MODULE=on go test -run XXX -bench .
