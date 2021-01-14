TARGET = main

all: $(TARGET)

$(TARGET): main.go provider.go session.go user.go db.go templates.go test.go device.go
	go build -o $(TARGET) .

go.sum:
	go get
	go get .

clean:
	$(RM) $(TARGET) go.sum
