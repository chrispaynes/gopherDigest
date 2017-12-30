src = ./src
main = $(src)/main.go
pkgDir = $(src)/$(pkg)

.PHONY: clean build dockerUp fmt install start test src-package SQLdata

build:
	docker-compose build
	docker-compose up

dockerUp: clean build
	@docker-compose down
	@docker-compose up

clean:
	@rm -f ./cmd/main

fmt: 
	@go fmt ./...

install:
	go install $(main)

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
	@go run $(main)

test:
	@go test **/*_test.go