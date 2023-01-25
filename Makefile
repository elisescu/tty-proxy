VERSION := $(shell git-semv now | sed -e 's/v//g')
registry   = pyama
repository = tty-proxy
image:
	docker build --platform linux/amd64 -t $(registry)/$(repository):$(VERSION) .
	docker push $(registry)/$(repository):$(VERSION)

.PHONY: releasedeps
releasedeps: git-semv

.PHONY: git-semv
git-semv:
ifeq ($(shell uname),Linux)
	which git-semv || (wget https://github.com/linyows/git-semv/releases/download/v1.2.0/git-semv_linux_x86_64.tar.gz && tar zxvf git-semv_linux_x86_64.tar.gz && sudo mv git-semv /usr/bin/)
else
	which git-semv > /dev/null || brew tap linyows/git-semv
	which git-semv > /dev/null || brew install git-semv
endif

## release_major: release nke (major)
release_major: releasedeps
	git semv major --bump

.PHONY: release_minor
## release_minor: release nke (minor)
release_minor: releasedeps
	git semv minor --bump

.PHONY: release_patch
## release_patch: release nke (patch)
release_patch: releasedeps
	git semv patch --bump
