src = ./src
main = $(src)/main.go
pkgDir = $(src)/$(pkg)

.PHONY: build clean coverage dockerUp fmt install perconaTools start test src-package SQLdata vet

build:
	docker-compose build

dockerUp: clean
	@docker-compose up -d
	@python -m webbrowser "http://localhost:8080" &> /dev/null

clean:
	@docker-compose down
	@rm -f ./main

coverage:
	@set -e;
	@echo "mode: set" > acc.out;

	@for Dir in $$(find . -type d); do \
		if ls "$$Dir"/*.go &> /dev/null; then \
			go test -coverprofile=profile.out "$$Dir"; \
			go tool cover -html=profile.out; \
		fi \
	done

	@rm -rf ./profile.out;
	@rm -rf ./acc.out;

fmt:
	@go fmt ./...

install:
	go install $(main)

perconaTools:
	if [ -n "$$(grep -Ei 'debian|ubuntu|mint' /etc/*release)" ]; then \
		wget "https://www.percona.com/downloads/percona-toolkit/3.0.5/binary/debian/stretch/x86_64/percona-toolkit_3.0.5-1.stretch_amd64.deb" \
		&& sudo apt install --no-install-recommends -y ./percona-toolkit_3.0.5-1.stretch_amd64.deb; \
	fi;
	if [ -n "$$(grep -Ei 'fedora|redhat' /etc/*release)" ]; then \
		wget "https://www.percona.com/downloads/percona-toolkit/3.0.5/binary/redhat/7/x86_64/percona-toolkit-3.0.5-1.el7.x86_64.rpm" \
		&& sudo dnf install -y ./percona-toolkit-3.0.5-1.el7.x86_64.rpm; \
	fi;

SQLdata:
	docker cp ./test_db/employees.sql MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_departments.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_dept_emp.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_dept_manager.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_employees.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_salaries1.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_salaries2.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_salaries3.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/load_titles.dump MySQL:/docker-entrypoint-initdb.d/ \
	&& docker cp ./test_db/objects.sql MySQL:/docker-entrypoint-initdb.d/ \
	&& docker exec MySQL sh -c 'mysql -uroot -p"$$MYSQL_ROOT_PASSWORD" < /docker-entrypoint-initdb.d/employees.sql'

src-package:
	@mkdir -p $(pkgDir)
	@echo package $(pkg) | tee $(pkgDir)/$(pkg).go $(pkgDir)/$(pkg)_test.go

start: clean
	go run $(main)

test:
	@go test **/*_test.go

vet:
	@go vet ./...