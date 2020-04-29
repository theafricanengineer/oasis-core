OASIS_GO_SRC="../go/"

all: oasis-node

oasis-node: oasis-node-metrics

oasis-node-metrics:
	oasis-node/genmetricsdocs.py ${OASIS_GO_SRC} oasis-node/metrics.md.tpl > oasis-node/metrics.md
