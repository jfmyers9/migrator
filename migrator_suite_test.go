package migrator_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoMigrator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoMigrator Suite")
}

var db *sql.DB

var _ = BeforeSuite(func() {
	var err error
	db, err = sql.Open(getDatabaseDriver(), getGlobalDatabaseConnectionString())
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Ping()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(db.Close()).To(Succeed())
})

var _ = BeforeEach(func() {
	_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE migratortest%d`, GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	_, err := db.Exec(fmt.Sprintf(`DROP DATABASE migratortest%d`, GinkgoParallelNode()))
	Expect(err).NotTo(HaveOccurred())
})

func getDatabaseDriver() string {
	driver := os.Getenv("TEST_SQL_DRIVER")
	if driver == "" {
		driver = "postgres"
	}

	return driver
}

func getGlobalDatabaseConnectionString() string {
	connectionString := os.Getenv("TEST_SQL_CONNECTION")
	if connectionString == "" {
		connectionString = "postgres://localhost:5432?sslmode=disable"
	}

	return connectionString
}

func getLocalDatabaseConnectionString() string {
	connectionString := os.Getenv("TEST_SQL_CONNECTION")
	if connectionString == "" {
		connectionString = fmt.Sprintf("postgres://localhost:5432?sslmode=disable&dbname=migratortest%d", GinkgoParallelNode())
	}

	return connectionString
}
