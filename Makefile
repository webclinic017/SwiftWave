# Variables:
DIST_DIR = swiftwave_service/dashboard/www
SUBMOD_DIR = dashboard
AGENT_DIR = agent

main: build_service

build_dashboard:
	npm run build:dashboard
	
build_service: | build_dashboard
	CGO_ENABLED=0 go build .
	
install: build_service
	cp swiftwave /usr/bin/swiftwave

build_agent:
	cd $(AGENT_DIR) && go build -o swiftwave-agent .

install_agent: build_agent 
	cp $(AGENT_DIR)/swiftwave-agent /usr/bin/swiftwave-agent