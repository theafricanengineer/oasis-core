OASIS_GO_ROOT="../go/"

all: oasis-node

oasis-node: oasis-node-metrics

oasis-node-metrics:
	oasis-node/genmetricsdocs.py ${OASIS_GO_ROOT} oasis-node/metrics.md.tpl > oasis-node/metrics.md
