all: main

cf_sources := $(wildcard *.go)

docs/swagger.json: $(cf_sources)
	~/go/bin/swag init

mains: $(cf_sources) docs/swagger.json
	go build -o main .
	touch mains

main: $(cf_sources)
	go build -o main .

go.sum:
	go get
	go get .

clean:
	$(RM) $(TARGET) go.sum
