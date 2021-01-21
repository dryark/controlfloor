TARGET = main

all: $(TARGET)

cf_sources := $(wildcard *.go)

$(TARGET): $(cf_sources)
	go build -o $(TARGET) .

go.sum:
	go get
	go get .

clean:
	$(RM) $(TARGET) go.sum
