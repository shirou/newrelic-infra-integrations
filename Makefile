.PHONY: build

CMDS=check_tcp

build:
	for cmd in $(CMDS) ; do \
	    mkdir -p build/$$cmd ; \
	    cd $$cmd ; go build -o ../build/$$cmd/$$cmd ; \
		cp *-config.yml.sample ../build/$$cmd/ ; \
		cp *-definition.yml ../build/$$cmd/ ; \
	done
