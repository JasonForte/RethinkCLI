package main

import (
	"errors"
	"fmt"
	"os"

	"flag"

	"strings"

	r "gopkg.in/gorethink/gorethink.v3"
)

func printConnOpts(connOpts r.ConnectOpts) {
	fmt.Println("Connecting to database:", connOpts.Address)
}

// Conn Return a connection session
func Conn(connOpts r.ConnectOpts) *r.Session {

	printConnOpts(connOpts)

	session, err := r.Connect(connOpts)

	if err != nil {
		fmt.Println("error creating connection to rethinkdb")
	}

	return session
}

func hasDatabase(session *r.Session, database string) (bool, error) {

	fmt.Printf("Checking for database %s ... ", database)

	// Check if the database exists.
	cursor, err := r.DBList().Contains(database).Run(session)
	if err != nil {
		println("error fetching database list")
		return false, errors.New("could not perform query")
	}

	defer cursor.Close()

	// result of the query
	var res bool
	cursor.One(&res)

	if !res {
		fmt.Println("Does not exist")
	} else {
		fmt.Println("Exists")
	}

	return res, nil
}

// Create a database with the given name.
func createDatabase(session *r.Session, database string) (string, error) {
	_, err := r.DBCreate(database).RunWrite(session)
	if err != nil {
		return "", errors.New("error creating database")
	}

	fmt.Println("Database Created:", database)
	return database, nil
}

// Ensure that the given database exists on the given connection
func ensureDatabase(session *r.Session, database string) (bool, error) {

	hasDb, err := hasDatabase(session, database)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	// If the database exists then return
	if hasDb {
		return true, nil
	}

	_, err = createDatabase(session, database)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Check that the given database has the given table.
func hasTable(session *r.Session, database string, table string) (bool, error) {

	fmt.Printf("Checking for table %s.%s ... ", database, table)

	cursor, err := r.DB(database).TableList().Contains(table).Run(session)

	if err != nil {
		return false, errors.New("could not stat tables in database. check database is created")
	}

	defer cursor.Close()

	var res bool
	cursor.One(&res)

	if !res {
		fmt.Println("Does not exist")
	} else {
		fmt.Println("Exists")
	}

	return res, nil

}

// Create a table in the given database with the given PK.
func createTable(session *r.Session, database string, table string, pk string) (string, error) {

	_, err := r.DB(database).TableCreate(table, r.TableCreateOpts{PrimaryKey: pk}).RunWrite(session)

	if err != nil {
		return "", errors.New("could not create table")
	}

	fmt.Println("Table created:", table)
	return table, nil
}

// If the given table is not in the given database - create it
func ensureTable(session *r.Session, database string, table string, pk string) (bool, error) {
	hasTb, err := hasTable(session, database, table)

	if err != nil {
		return false, err
	}

	if hasTb {
		return true, nil
	}

	_, err = createTable(session, database, table, pk)
	if err != nil {
		return false, err
	}

	return true, nil
}

func helpText() {
	fmt.Println(`
	Commands
  	---------
	ensure_database - Ensure that a rethinkdb database exists.
  	ensure_table    - Ensure that a table is in the given database.
	`)
}

func cmdConnOpts() r.ConnectOpts {

	var host string
	flag.StringVar(&host, "host", "localhost", "specify the hostname of the rethinkdb node. defaults to localhost")

	var port int
	flag.IntVar(&port, "port", 28015, "specify rethinkdb port. defaults to 28015")

	flag.Parse()

	fmt.Println("RethinkDB:", fmt.Sprintf("%s:%d", host, port))

	return r.ConnectOpts{
		Address: fmt.Sprintf("%s:%d", host, port),
	}

}

func cmdEnsureDatabase(connOpts r.ConnectOpts) (int, error) {
	fmt.Println("ensure database")

	sl := flag.Args()
	param := sl[len(sl)-1]

	conn := Conn(connOpts)
	defer conn.Close()

	_, err := ensureDatabase(conn, param)
	if err != nil {
		return 1, err
	}

	return 0, nil
}

func cmdEnsureTable(connOpts r.ConnectOpts) (int, error) {
	fmt.Println("ensure table")

	sl := flag.Args()
	params := strings.Split(sl[len(sl)-1], ".")

	conn := Conn(connOpts)
	defer conn.Close()

	if len(params) < 2 {
		return 1, errors.New("need table in form {db}.{table}.{pk}")
	} else if len(params) == 2 {
		params = append(params, "id")
	} else if len(params) > 3 {
		return 1, errors.New("invalid param for create_table")
	}

	_, err := ensureTable(conn, params[0], params[1], params[2])
	if err != nil {
		return 1, err
	}

	return 0, nil
}

func getCommand() func(r.ConnectOpts) (int, error) {

	for _, a := range os.Args[1:] {
		switch a {
		case "ensure_database":
			return cmdEnsureDatabase
		case "ensure_table":
			return cmdEnsureTable
		default:
			continue
		}
	}

	return func(_ r.ConnectOpts) (int, error) {
		helpText()
		return 1, errors.New("unknown action")
	}
}

func main() {

	// first param is the application name
	cmd := os.Args[0]
	fmt.Printf("Program name: %s\n", cmd)

	argCount := len(os.Args[1:])

	if argCount == 0 {
		helpText()
		os.Exit(1)
	}

	connOpts := cmdConnOpts()

	command := getCommand()

	res, err := command(connOpts)

	if err != nil {
		fmt.Println(err)
	}

	os.Exit(res)

}
