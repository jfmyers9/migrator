# migrator

A simple library for running migrations against a SQL database.

## Usage

```bash
go get -u -v github.com/jfmyers9/migrator
```

```golang
migrator, err := migrator.New(db, migrations...)
if err != nil {
  return err
}

err = migrator.Migrate()
if err != nil {
  return err
}
```

```golang
type CreateUsersTable struct {}

func (m *CreateUsersTable) Name() string {
  return "create_uers_table"
}

func (m *CreateUsersTable) Version() int {
  return 1581991825
}

func (m *CreateUsersTable) Up(tx *sql.Tx) error {
  _, err := tx.Exec(`CREATE TABLE users (
    username VARCHAR (50) UNIQUE NOT NULL,
    password_salt VARCHAR (255) NOT NULL,
    email VARCHAR (50) NOT NULL
    created_at TIMESTAMP NOT NULL
  )`)
  if err != nil {
    return err
  }
}

func (m *CreateUsersTable) Down(tx *sql.Tx) error {
  //TODO: Rollback not implemented yet.
  return nil
}
```
