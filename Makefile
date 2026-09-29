.PHONY:all long

all:

long:
	# takes half hour maybe. check Go versus Java SANY parse by comparing XML output
	GOCACHE=$(CURDIR)/.codex-gocache \
	GOTMPDIR=$(CURDIR)/.codex-gotmp \
	TLAGO_SANY_XML_CORPUS_ARTIFACT_DIR=.codex-sany-artifacts \
	TLAGO_SANY_XML_CORPUS=1 go test -run TestSanyXMLTargetCorpusAgainstJavaSANY ./
