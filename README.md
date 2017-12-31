# gopherDigest
gopherDigest is an experimental playground to test MySQL queries using [Percona Toolkit's PT-Query-Digest tool](https://www.percona.com/doc/percona-toolkit/LATEST/pt-query-digest.html) to Analyze MySQL queries from logs, processlist, and tcpdump.

# Status
Work in Progress

## Quickstart
- Download [Docker](https://www.docker.com/) and Docker-Compose
- Download the [MySQL Employees DB on GitHub](https://github.com/datacharmer/test_db)
- Copy the `test_db` directory into the root of the `gopherDigest` directory
- Modify the bottom of the `test_db/employees.sql` script so that the .dump source files are prefixed with `/docker-entrypoint-initdb.d/`, for example `source load_departments.dump;` should be `source /docker-entrypoint-initdb.d/load_departments.dump ;` Remove `source show_elapsed.sql ;` from the end of the file, as we will not need it.
- Run `docker pull mysql:8.0` so that the image will be ready to accept SQL data
- Run `makeSQLdata` to copy necessary SQL scripts from the `test_db` directory into the MySQL image
- Using the `gopherDigest/Docker/mysql.TEMPLATE.env` template, populate the Environment Variables and save the file as `gopherDigest/Docker/mysql.env`. To run the app on the host machine, you will need to export the environment variables to the shell
- Run `make build` to spin up the Docker GopherDigest, and MySQL Server application services


## Configuration
The following tables lists the configurable application environment variables that need to be defined in the `gopherDigest/Docker/mysql.env`.

| Parameter        | Description           | Example  |
| ------------- |-------------| -----|
| MYSQL_USER | These variables are optional, used in conjunction to create a new user and to set that user's password. This user will be granted superuser permissions (see above) for the database spPecified by the MYSQL_DATABASE variable. *Both variables are required for a user to be created. | john |
| MYSQL_PASSWORD      | These variables are optional, used in conjunction to create a new user and to set that user's password. This user will be granted superuser permissions (see above) for the database specified by the MYSQL_DATABASE variable. *Both variables are required for a user to be created. | johnPassword123 |
| MYSQL_HOST      | Host name or IP Address used to create a TCP/IP connection to the MySQL Server. | 127.0.0.1, localhost |
| MYSQL_PORT      | The port number to use for the connection, for connections made using TCP/IP. T | 3306 |
| MYSQL_DATABASE      | This variable is optional and allows you to specify the name of a database to be created on image startup. If a user/password was supplied then that user will be granted superuser access (corresponding to GRANT ALL) to this database. | employees |
| MYSQL_ROOT_PASSWORD      | This variable is mandatory and specifies the password that will be set for the MySQL root superuser account. | secretRootPassword |
| MYSQL_SOCKET      | MySQL data transmission socket file location as defined in /etc/my.cnf. | /var/run/mysqld/mysqld.sock |