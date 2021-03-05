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

benchmark: bins
	@GO111MODULE=on go test -run XXX -bench .
