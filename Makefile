.PHONY: all

all: clean prep build format

clean:
	rm -rf kurento

prep:
	mkdir kurento

build:
	go run main.go
	# Because I don't find os.ModeXXX to use when I create files...
	chmod -R a-x kurento/*.go

format:
	cp kurento_go_base/*.go kurento/
	goimports -w ./kurento

test:
	cd kurento && go test -v
