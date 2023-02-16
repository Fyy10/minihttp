demo:
	go run cmd/demo/main.go

test:
	go test .

benchmark:
	go test -bench=.
