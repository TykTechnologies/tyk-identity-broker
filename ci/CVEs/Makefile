PREFIX  	:= tykio
SCAN_REPORTS 	:= $(addsuffix .json,tyk-gateway tyk-dashboard tyk-plugin-compiler tyk-hybrid-docker tyk-pump tyk-mdcb-docker tyk-identity-broker)
SCAN 		:= -docker scan --accept-license --json --group-issues
# Any jq filter that returns an array can be used here
FILTER 	:= .vulnerabilities | group_by(.packageName)[] | add | [ .packageName, .severity, .severityWithCritical, .version, .isUpgradeable, .isPatchable, .description, .from[0][0], .identifiers.CVE[] ]

all: unified.csv

unified.csv: $(SCAN_REPORTS)
	jq -r '$(FILTER) | @csv' $^ > $@

tyk-gateway.json tyk-dashboard.json tyk-plugin-compiler.json tyk-hybrid-docker.json:
	$(SCAN) $(PREFIX)/$(basename $@):v3.2 > $@

tyk-pump.json:
	$(SCAN) $(PREFIX)/$(basename $@):v1.5 > $@

tyk-mdcb-docker.json:
	$(SCAN) $(PREFIX)/$(basename $@):v1.7 > $@

tyk-identity-broker.json:
	$(SCAN) $(PREFIX)/$(basename $@):v1.2 > $@

clean:
	rm -fv *.csv

realclean: clean
	rm -fv *.json