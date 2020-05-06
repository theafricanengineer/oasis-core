OASIS_GO_SRC="../go/"

all: update

update:
	../scripts/generate-metrics-docs.py ${OASIS_GO_SRC} oasis-node/metrics.md.tpl > oasis-node/metrics.md

.PHONY:
	all
	update
