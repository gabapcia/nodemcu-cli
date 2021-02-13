BUILDPATH=$(CURDIR)/bin
GO=$(shell which go)

EXENAME=nodemcu

install: build
	@mv $(BUILDPATH)/$(EXENAME) /usr/bin
	@$(MAKE) clean

clean:
	@rm -rf $(BUILDPATH)

build:
	@if [ ! -d $(BUILDPATH) ] ; then mkdir -p $(BUILDPATH) ; fi
	@$(GO) build -v -o $(BUILDPATH)/$(EXENAME) main.go
