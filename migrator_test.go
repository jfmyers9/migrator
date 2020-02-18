package migrator_test

import (
	"database/sql"
	"errors"

	"github.com/jfmyers9/migrator"
	"github.com/jfmyers9/migrator/migratorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/lib/pq"
)

var _ = Describe("Migrator", func() {
	var (
		m          *migrator.Migrator
		migration1 *migratorfakes.FakeMigration
		migration2 *migratorfakes.FakeMigration
		db         *sql.DB
	)

	BeforeEach(func() {
		var err error

		db, err = sql.Open(getDatabaseDriver(), getLocalDatabaseConnectionString())
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Ping()).To(Succeed())

		migration1 = &migratorfakes.FakeMigration{}
		migration1.NameReturns("migration1")
		migration1.VersionReturns(1)

		migration2 = &migratorfakes.FakeMigration{}
		migration2.NameReturns("migration2")
		migration2.VersionReturns(2)

		m = migrator.NewMigrator(db, migration1, migration2)
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	Context("Setup", func() {
		It("creates the migrations table if it does not already exist", func() {
			err := m.Setup()
			Expect(err).NotTo(HaveOccurred())

			rows, err := db.Query("SELECT count(*) FROM schema_migrations")
			Expect(err).NotTo(HaveOccurred())

			Expect(rows.Next()).To(BeTrue())

			var count int64
			err = rows.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(BeEquivalentTo(0))

			Expect(rows.Next()).To(BeFalse())
		})

		Context("when the schema_migrations table already exists", func() {
			BeforeEach(func() {
				_, err := db.Exec("CREATE TABLE schema_migrations ( name VARCHAR (50) NOT NULL, version INTEGER UNIQUE NOT NULL )")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not error", func() {
				err := m.Setup()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("Migrate", func() {
		BeforeEach(func() {
			Expect(m.Setup()).To(Succeed())
		})

		It("runs the migrations in order", func() {
			err := m.Migrate()
			Expect(err).NotTo(HaveOccurred())
			Expect(migration1.UpCallCount()).To(Equal(1))
			Expect(migration2.UpCallCount()).To(Equal(1))

			rows, err := db.Query("SELECT name, version FROM schema_migrations")
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()

			var name1, name2 string
			var version1, version2 int64
			Expect(rows.Next()).To(BeTrue())
			Expect(rows.Scan(&name1, &version1)).To(Succeed())
			Expect(name1).To(Equal(migration1.Name()))
			Expect(version1).To(BeEquivalentTo(migration1.Version()))

			Expect(rows.Next()).To(BeTrue())
			Expect(rows.Scan(&name2, &version2)).To(Succeed())
			Expect(name2).To(Equal(migration2.Name()))
			Expect(version2).To(BeEquivalentTo(migration2.Version()))

			Expect(rows.Next()).To(BeFalse())
		})

		Context("when previous migrations have already been run", func() {
			BeforeEach(func() {
				_, err := db.Exec("INSERT INTO schema_migrations VALUES ($1, $2)", migration1.Name(), migration1.Version())
				Expect(err).NotTo(HaveOccurred())
			})

			It("skips any already applied migrations", func() {
				err := m.Migrate()
				Expect(err).NotTo(HaveOccurred())
				Expect(migration1.UpCallCount()).To(Equal(0))
				Expect(migration2.UpCallCount()).To(Equal(1))

				rows, err := db.Query("SELECT name, version FROM schema_migrations")
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				var name1, name2 string
				var version1, version2 int64
				Expect(rows.Next()).To(BeTrue())
				Expect(rows.Scan(&name1, &version1)).To(Succeed())
				Expect(name1).To(Equal(migration1.Name()))
				Expect(version1).To(BeEquivalentTo(migration1.Version()))

				Expect(rows.Next()).To(BeTrue())
				Expect(rows.Scan(&name2, &version2)).To(Succeed())
				Expect(name2).To(Equal(migration2.Name()))
				Expect(version2).To(BeEquivalentTo(migration2.Version()))

				Expect(rows.Next()).To(BeFalse())
			})
		})

		Context("when one of the migration fails", func() {
			BeforeEach(func() {
				migration1.UpReturns(errors.New("boom"))
			})

			It("doesn't run the following migrations and doesn't commit anything to the database", func() {
				err := m.Migrate()
				Expect(err).To(HaveOccurred())
				Expect(migration1.UpCallCount()).To(Equal(1))
				Expect(migration2.UpCallCount()).To(Equal(0))

				rows, err := db.Query("SELECT name, version FROM schema_migrations")
				Expect(err).NotTo(HaveOccurred())
				defer rows.Close()

				Expect(rows.Next()).To(BeFalse())
			})
		})
	})
})
