package main

import (
	"devops-tools.pearson.com/mysp/cassandra-migrate/cql"
	"flag"
	"fmt"
	"github.com/gocql/gocql"
	"os"
	"sort"
	"strings"
)

var (
	dryRun   = true
	confFile = "./conf/example.toml"
	env      = "local"
)

func init() {
	flag.StringVar(&confFile, "conf", confFile, "Configuration file path")
	flag.BoolVar(&dryRun, "dryrun", dryRun, "Dry Run")
	flag.StringVar(&env, "env", env, "Environment against which to run migrations")
}

func main() {
	flag.Parse()

    action := ""
	args := flag.Args()
	if len(args) > 0 {
		action = args[0]
	}

	conf, confErr := cql.NewMigrationConfig(confFile)
	if confErr != nil {
		fail("Failed to read configuration file: '%s'", confFile)
	}

	// TODO: We're not currently guarding against any sort of odd environment names here or in the config. Should we?
	//       At some point the user has to take responsibility for their own choices, right?
	//env = strings.ToLower(env)

	fmt.Printf("Environment: %s\n", env)
	fmt.Printf("Cassandra Seed Node: %s\n", conf.Environments[env].CassandraHosts)
	fmt.Printf("Migration Scripts Path: %s\n", conf.Scripts.Path)
    fmt.Printf("DRY RUN?: %v\n", dryRun)

	switch action {
	case "log":
		listLog(conf, env)

	case "up":
		fmt.Printf("Migrate %s...\n", action)
		up(dryRun, conf, env)

	case "create":
		if len(args) < 2 {
			fail("Migrate %s \"Missing migration name\"", action)
		}

		var createErr error
		if len(args) > 2 {
			createErr = create(conf, args[1], args[2])
		} else {
			createErr = create(conf, args[1], "")
		}
		if createErr != nil {
			fail("Unable to create migration file", createErr)
		}

	default:
		fmt.Printf("Unknown command \"%s\"\n", action)
		flag.Usage()
		os.Exit(1)
	}
}

func fail(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

// TODO: Need to figure out whether we need to provide a cluster or whether one of the 'seeds' is ok.
func initDB(h, k string) (*gocql.Session, error) {
    fmt.Printf("Connecting to %s/%s\n", h, k)
    cluster := gocql.NewCluster(h)
    cluster.Keyspace = k
    cluster.Consistency = gocql.Quorum
    cluster.ConnPoolType = gocql.NewSimplePool
    return cluster.CreateSession()
}

func initSchemaVersion(session *gocql.Session) error {
    schemaVerCQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS schema_version(
                    applied timestamp,
                    environment text,
                    name text,
                    checksum blob,
                    user text,
                    version text,
                    PRIMARY KEY (name, version)) WITH CLUSTERING ORDER BY (version ASC)`)

    iter := session.Query(schemaVerCQL).Iter()
    return iter.Close()
}

//
// Actually apply the CQL statements to the DB. TODO: Actually apply the CQL statements to the DB ;)
//
func applyUpdate(migration *cql.Migration, filename string, session *gocql.Session) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Printf("Applying migration: %s\n   |%-40s|%-15s|%-12s|%-40x|\n",
		migration.File,
		migration.Name,
		migration.Environment,
		migration.Version,
		migration.Sum)

	statements, err := cql.ReadCQLFile(f)
	if err != nil {
		return err
	}

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if len(statement) > 0 {
			//fmt.Printf("Applying %d: %s \n", i, statement)
		}
	}
	if saveErr := migration.Save(session); saveErr != nil {
		return saveErr
	}
	return nil
}

//
// Just spit out the content of the schema_version table for all to see.
//
func listLog(conf *cql.MigrationConfig, env string) {
	hosts := conf.Environments[env].CassandraHosts
	keyspace := conf.Environments[env].Keyspace

	session, err := initDB(hosts, keyspace)
	if err != nil {
		fail("Failed to connect: %s - %q", hosts, err)
	}
	defer session.Close()

	applied := cql.ListAppliedMigrations(session)
	fmt.Println("Previously Applied Migrations:")
	fmt.Printf("    |%-40s|%-20s|%-15s|%-20s|%-20s|\n", "Name", "Version", "Environment", "Applied By", "Applied On")
	for _, a := range applied {
		fmt.Printf("    |%-40s|%-20s|%-15s|%-20s|%-20s|\n", a.Name, a.Version, a.Environment, a.User, a.Applied)
	}
}

//
// Function to create a new migration *file* with the correct name formatting etc such
// that the user may add her CQL to it.
//
func create(conf *cql.MigrationConfig, name string, env string) error {
	m := cql.CreateMigration(name, env)
	if err := m.CreateMigrationFile(conf.Scripts.Path); err != nil {
		return fmt.Errorf("Failed to create migration file: %s", err.Error())
	}
	return nil
}

//
// Migrate up. We don't do down yet. Need to do a better job of parsing the CQL to do that, I think.
//
func up(dryRun bool, conf *cql.MigrationConfig, env string) {
	hosts := conf.Environments[env].CassandraHosts
	keyspace := conf.Environments[env].Keyspace

	session, err := initDB(hosts, keyspace)
	if err != nil {
		fail("Failed to connect: %s - %q", hosts, err)
	}
	defer session.Close()

	if !dryRun {
		err = initSchemaVersion(session)
		if err != nil {
			fail("Failed to init schema: %q", err)
		}
	}
	// Retrieve all the previously applied updates from the DB.
	applied := cql.ListAppliedMigrations(session)

	// Create Migration objects from each candidate file in the specified scripts path.
	updates, listErr := cql.ListMigrationFiles(conf.Scripts.Path)
	if listErr != nil {
		fail("Failed to list migration files", err.Error())
	}

	// Ensure the files are in version order
	sort.Sort(updates)

	// Run each update if:
	//   1. It is not detected in the 'applied' list from above.
	//   2. The dry run flag was not set.
	for _, m := range updates {
		if applied.Contains(m) {
			fmt.Printf("x: '%s'\n", m.File)
		} else {
			if !dryRun {
				if err := applyUpdate(m, m.File, session); err != nil {
					fail(err.Error())
				}
			}
		}
	}
}
