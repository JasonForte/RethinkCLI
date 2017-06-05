# RethinkCLI
CLI tool for RethinkDB

## Overview

This tool allows an automated script to ensure that a database / table exists before booting a microservice.

## Runninng the Tool

 1. Clone this repo
 2. Run `make`


## Example Usage

This tool is for use with RethinkDB

```sh
$ # Run rethinkdb instance with docker
$ docker run -d -p 28015:28015 -p 8080:8080 rethinkdb
```

### Ensure that a Database exists

```sh
$ # check that a database exists and create if it doesn't
$ ./dbtool -host localhost -port 28015 ensure_database random_database && echo "DATABASE IS PRESENT!"
```

### Ensure that a Table exists

If the table doesn't exist then the last element will be the PK of the table.

```sh
$ # check that a table exists and if not create it.
$ ./dbtool -host localhost -port 28015 ensure_table random_database.random_table.CUSTOM_PK 
```

