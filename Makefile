# Main Makefile for surv-export

VPATH=  surv-export

all: surv-export

clean:
	rm -f surv-export

install:
	go install -v

surv-export:    surv-export.go cli.go
	go build -v -o surv-export

push:
	git push --all
	git push --tags
	git push --all backup
	git push --tags backup
	git push --all gitlab
	git push --tags gitlab
