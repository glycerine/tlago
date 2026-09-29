.PHONY:all long

all:

long:
	# takes about 5-7 minutes. check Go versus Java SANY parse by comparing XML output
	# A pre-requisite: "cd test_vectors && make expected-xml" may be needed
	# to have the Java SANY generate .xml.gold for comparison for each spec. These
	# xml files are too big in total to commit to the repo (922 MB currently).
	mkdir .codex-gotmp || true
	GOCACHE=$(CURDIR)/.codex-gocache \
	GOTMPDIR=$(CURDIR)/.codex-gotmp \
	TLAGO_SANY_XML_CORPUS_ARTIFACT_DIR=.codex-sany-artifacts \
	TLAGO_SANY_XML_CORPUS=1 go test -v -run TestSanyXMLTargetCorpusAgainstJavaSANY -timeout=0 -count=1 ./
