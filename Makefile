BUILDPATH=$(CURDIR)/bin
GO=$(shell which go)

EXENAME=nodemcu


install: makedir build global clean


makedir:
	@if [ ! -d $(BUILDPATH) ] ; then mkdir -p $(BUILDPATH) ; fi

clean:
	@rm -rf $(BUILDPATH)

build:
	@$(GO) build -v -o $(BUILDPATH)/$(EXENAME) main.go

global:
	@mv $(BUILDPATH)/$(EXENAME) /usr/bin
