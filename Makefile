.PHONY: all

all: clean prep build format


clean:
	rm -rf kurento

prep:
	mkdir kurento

build:
	go run main.go

format:
	cp kurento_go_base/*_test.go kurento/
	go fmt ./kurento
