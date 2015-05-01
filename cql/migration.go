package cql

import (
	"crypto/sha1"
	"fmt"
	"github.com/gocql/gocql"
	"golang.org/x/text/unicode/norm"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	migrationTimeFormat = "200601021504"
)

type Migration struct {
	Applied     time.Time
	Environment string
	Name        string
	Sum         []byte
	User        string
	Version     string
	File        string
}

func CreateMigration(name string, env string) (migration *Migration) {
	var environment string = "all"

	if len(env) > 0 {
		environment = env
	}
	name = sanitizeStr(name)

	migration = &Migration{
		Environment: environment,
		Name:        name,
		Version:     time.Now().Format(migrationTimeFormat),
	}
	return migration
}

//
// Create a Migration object based upon a file's name. The date stamp (or 'version')
// of the file is expected in the format 'YYYYMMDDhhmm'.
//
func MigrationFromFile(path string) (migration *Migration, err error) {
	fbytes, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}

	filename := filepath.Base(path)
	re := regexp.MustCompile("(\\d{12})[_.]([A-Za-z0-9_-]+)\\.([A-Za-z0-9_-]+)\\.cql")

	cksum := sha1.New()
	cksum.Write(fbytes)
	filesha1 := cksum.Sum(nil)

	if ok := re.MatchString(filename); ok {
		matcher := re.FindAllStringSubmatch(filename, -1)
		migration = &Migration{
			Environment: matcher[0][3],
			Name:        matcher[0][2],
			Sum:         filesha1,
			Version:     matcher[0][1],
			User:        currentUser(),
			File:        path,
		}
	} else {
		return nil, fmt.Errorf("File did not match the expected naming convention")
	}
	return migration, nil
}

//
// With a directory path (string), create a Migration object for each "*.cql" file
// in that directory and return a map of Migrations->File paths
//
// Returns a slice of Errors in the event that individual files cause problems. I
// hope that this will enable us to catch all the problematic files in one go, not
// one at a time.
//
func ListMigrationFiles(path string) (updates Migrations, errs Errors) {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		errs = append(errs, err)
		return updates, errs
	}

	for _, f := range files {
		if f.Mode().IsRegular() && strings.HasSuffix(f.Name(), ".cql") {
			fpath := path + "/" + f.Name()

			if m, err := MigrationFromFile(fpath); err == nil {
				updates = append(updates, m)
			} else {
				errs = append(errs, fmt.Errorf("Failed to create migration from file: '%s': %s", fpath, err.Error()))
			}
		}
	}
	if len(errs) == 0 {
		return updates, nil
	}
	return updates, errs
}

//
// Pull out all of the Migrations that the schema_version table knows about.
//
func ListAppliedMigrations(session *gocql.Session) (applied Migrations) {
	iter := session.Query(`SELECT applied, environment, checksum, name, user, version FROM schema_version`).Iter()

	for update := new(Migration); iter.Scan(&update.Applied, &update.Environment, &update.Sum, &update.Name, &update.User, &update.Version); update = new(Migration) {
		applied = append(applied, update)
	}
	return applied
}

func sanitizeStr(v string) string {
	var whiteSpace = regexp.MustCompile("[^\\w]+")
	return strings.ToLower(whiteSpace.ReplaceAllString(norm.NFKD.String(v), "_"))
}

func currentUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	} else if u.Name != "" {
		return u.Name
	}
	return u.Username
}

//
// Create a file into which the user can put her CQL. It has the correctly formatted
// timestamp (we call it 'version') and a name that has been sanitised of unicode
// and whitespace characters. It also contains an 'environment' name.
//
func (m *Migration) CreateMigrationFile(dirPath string) error {
	m.File = dirPath + "/" + m.Version + "_" + m.Name + "." + m.Environment + ".cql"
	fmt.Printf("Migrate Creating migration: '%s' in '%s'\n", m.Name, m.File)

	f, err := os.Create(m.File)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Chmod(0644)

	return nil
}

//
// Apply the CQL statements in the Migration file to the db specified by 'session'.
// Will attempt to apply all the statements and return a list of errors at the end.
// TODO: Is that actually a good idea!?
//
func (m *Migration) Apply(session *gocql.Session) (errs Errors) {
	f, fopenErr := os.Open(m.File)
	if fopenErr != nil {
		errs = append(errs, fopenErr)
		return errs
	}
	defer f.Close()

	fmt.Printf("Applying migration: %s\n   |%-40s|%-15s|%-12s|%-40x|\n",
		m.File,
		m.Name,
		m.Environment,
		m.Version,
		m.Sum)

	statements, readErr := ReadCQLFile(f)
	if readErr != nil {
		errs = append(errs, readErr)
		return errs
	}

	for _, st := range statements {
		st = strings.TrimSpace(st)
		if "" == st {
			continue
		}
		query := session.Query(st)
		if execErr := query.Exec(); nil != execErr {
			errs = append(errs, execErr)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

//
// Insert a record into the schema_version table for this Migration object.
//
func (m *Migration) Save(session *gocql.Session) error {
	saveCql := fmt.Sprintf(`
			INSERT INTO schema_version (
			            applied,
	                    environment,
	                    checksum,
	                    name,
	                    user,
	                    version)
			    VALUES( ?, ?, ?, ?, ?, ? )`)

	saveQuery := session.Query(saveCql, time.Now(), m.Environment, m.Sum, m.Name, m.User, m.Version)
	if queryErr := saveQuery.Exec(); nil != queryErr {
		return fmt.Errorf("Unable to save migration '%s': %s", m.Name, queryErr.Error())
	}
	return nil
}

func (m *Migration) Compare(other *Migration) bool {
	if m.Name == other.Name &&
		m.Version == other.Version &&
		m.Environment == other.Environment {
		return true
	}
	return false
}
