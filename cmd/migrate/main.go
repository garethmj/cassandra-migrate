package main

import (
	"devops-tools.pearson.com/mysp/cassandra-migrate/cql"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/gocql/gocql"
	"os"
	"sort"
)

var (
	app = kingpin.New("cassandra-migration", "Migrations tool for cassandra")

	// Main flags which must be provided before the commands below.
	dryRun   = app.Flag("dryrun", "Dry run").Short('d').Bool()
	confPath = app.Flag("conf", "Path to config file.").Short('c').Default("./conf/example.toml").String()
	env      = app.Flag("env", "Set config environment.").Short('e').Default("local").String()

	// The main commands.
	cmdCreate = app.Command("create", "Create new migration.")
	cmdList   = app.Command("list", "List all candidate migrations.")
    cmdLog    = app.Command("log", "List all applied migrations.")
	cmdUp     = app.Command("up", "Apply a first new migration.")

	// Options to the 'create' command.
	migrationName = cmdCreate.Arg("name", "Name of new migration.").Required().String()
	migrationEnv  = cmdCreate.Flag("target-env", "Name of new migration.").Short('t').Default("all").String()

	command = kingpin.MustParse(app.Parse(os.Args[1:]))
)

var (
	conf *cql.MigrationConfig
)

func main() {
	// Load dat config.
	conf = mustLoadConfig()

	// TODO: We're not currently guarding against any sort of odd environment names here or in the config. Should we?
	//       At some point the user has to take responsibility for their own choices, right?
	//env = strings.ToLower(*env)
	fmt.Printf("Environment: %s\n", *env)
	fmt.Printf("Cassandra Seed Node: %s\n", conf.Environments[*env].CassandraHosts)
	fmt.Printf("Migration Scripts Path: %s\n", conf.Scripts.Path)
	fmt.Printf("DRY RUN?: %v\n", *dryRun)

	switch command {

    case cmdList.FullCommand():
        listCandidates(conf, *env)

	case cmdLog.FullCommand():
		listLog(conf, *env)

	case cmdUp.FullCommand():
		fmt.Printf("Migrate up\n")
		up(*dryRun, conf, *env)

	case cmdCreate.FullCommand():
		if createErr := create(conf, *migrationName, *migrationEnv); createErr != nil {
			fail("Unable to create migration file", createErr)
		}

	default:
		app.Usage(os.Stdout)
	}
}

func mustLoadConfig() *cql.MigrationConfig {
	conf, confErr := cql.NewMigrationConfig(*confPath)
	if confErr != nil {
		fail("Failed to read configuration file: '%s'", *confPath)
	}
	return conf
}

// TODO: Need to figure out whether we need to provide a cluster or whether one of the 'seeds' is ok.
// TODO: Quorum or LocalQuorum? That is the question.
func mustConnectToDB(conf *cql.MigrationConfig, env string) *gocql.Session {
	hosts := conf.Environments[env].CassandraHosts
	ks := conf.Environments[env].Keyspace

	fmt.Printf("Connecting to %s/%s\n", hosts, ks)
	cluster := gocql.NewCluster(hosts)
	cluster.Keyspace = ks
	cluster.Consistency = gocql.Quorum
	cluster.ConnPoolType = gocql.NewSimplePool

	session, err := cluster.CreateSession()
	if err != nil {
		fail("Failed to connect: %s - %q", hosts, err)
	}
	return session
}

func fail(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
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
// Just spit out the content of the schema_version table for all to see.
//
func listLog(conf *cql.MigrationConfig, env string) {
	session := mustConnectToDB(conf, env)
	defer session.Close()

	applied := cql.ListAppliedMigrations(session)
	fmt.Println("Previously Applied Migrations:")
	fmt.Printf("    |%-40s|%-20s|%-15s|%-20s|%-20s\n", "Name", "Version", "Environment", "Applied By", "Applied On")
	for _, a := range applied {
		fmt.Printf("    |%-40s|%-20s|%-15s|%-20s|%-20s\n", a.Name, a.Version, a.Environment, a.User, a.Applied)
	}
}

//
// TODO: Err....so there's a lot of copy paste here from the up() fn.
//
func listCandidates(conf *cql.MigrationConfig, env string) {
    session := mustConnectToDB(conf, env)
    defer session.Close()

    applied := cql.ListAppliedMigrations(session)

    updates, listErr := cql.ListMigrationFiles(conf.Scripts.Path)
    if listErr != nil {
        fail("Failed to list migration files: %s", listErr.Error())
    }

    fmt.Println("Migration Candidates:")
    fmt.Printf("    |%-40s|%-20s|%-15s|%-11s|%-50s\n", "Migration Name", "Version", "Environment", "Candidate?", "File Path")

    for _, m := range updates {
        isCandidate := "yes"
        if m.Environment != "all" && m.Environment != env { isCandidate = "no" }

        if applied.Contains(m) {
            fmt.Printf("    |%-40s|%-20s|%-15s|%-11s|%-50s\n", m.Name, m.Version, m.Environment, isCandidate, m.File)
        } else {
            fmt.Printf("    |%-40s|%-20s|%-15s|%-11s|%-50s\n", m.Name, m.Version, m.Environment, isCandidate, m.File)
        }
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
	session := mustConnectToDB(conf, env)
	defer session.Close()

	if !dryRun {
		initErr := initSchemaVersion(session)
		if initErr != nil {
			fail("Failed to init schema: %q", initErr)
		}
	}
	// Retrieve all the previously applied updates from the DB.
	applied := cql.ListAppliedMigrations(session)

	// Create Migration objects from each candidate file in the specified scripts path.
	updates, listErr := cql.ListMigrationFiles(conf.Scripts.Path)
	if listErr != nil {
		fail("Failed to list migration files: %q", listErr)
	}

	// Ensure the files are in version order so they're applied in order.
	sort.Sort(updates)

	// Run each update if:
	//   1. It is not detected in the 'applied' list from above.
	//   2. The environment of the migration is not either 'all' or the same as the env flag.
	//   3. The dry run flag was not set.
	for _, m := range updates {
		if applied.Contains(m) {
			fmt.Printf("Ignoring: '%s' (already applied)\n", m.File)
		} else {
			if m.Environment != "all" && m.Environment != env {
				fmt.Printf("Ignoring: '%s' (because environment is '%s')\n", m.File, m.Environment)
				continue
			}
			if !dryRun {
				if err := m.Apply(session); err != nil {
					fail("Unable to apply migration '%s':\n   %s", m.Name, err.Error())
				}
				if err := m.Save(session); err != nil {
					fail("Unable to save migration '%s':\n   %s", m.Name, err.Error())
				}
			}
		}
	}
}
