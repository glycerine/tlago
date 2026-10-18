.PHONY: expected expected-xml expected-air

REGENERATE ?= 0
SPECS ?=

expected: expected-xml expected-air

expected-xml:
	REGENERATE="$(REGENERATE)" SPECS="$(SPECS)" bash ./generate_goldens.sh xml

expected-air:
	REGENERATE="$(REGENERATE)" SPECS="$(SPECS)" bash ./generate_goldens.sh air
