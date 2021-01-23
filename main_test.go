package main

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/ory/dockertest/docker"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

const (
	password = "123"
	database = "postgres"
)

type ExampleTestSuite struct {
	suite.Suite
	db          *sql.DB
	dbContainer *testPostgresContainer
}

type testPostgresContainer struct {
	databaseURL string
	pool        *dockertest.Pool
	resource    *dockertest.Resource
}

func (s *ExampleTestSuite) SetupSuite() {
	// Create & run
	postgresContainer, err := newRunningPostgresContainer()
	require.NoError(s.T(), err)

	// Connect
	db, err := postgresContainer.Connect()
	require.NoError(s.T(), err)

	// Migrate database
	m, err := migrate.New(
		"file://db/migrations",
		postgresContainer.databaseURL)
	require.NoError(s.T(), err, "failed to migrate db")
	require.NoError(s.T(), m.Up())

	s.db = db
	s.dbContainer = postgresContainer
}

func (s *ExampleTestSuite) TearDownSuite() {
	assert.NoError(s.T(), s.db.Close())
	assert.NoError(s.T(), s.dbContainer.Close(), "could not purge resource")
}

func newRunningPostgresContainer() (*testPostgresContainer, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11.10",
		Env:        []string{"POSTGRES_PASSWORD=" + password, "POSTGRES_DB=" + database},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, err
	}
	databaseURL := fmt.Sprintf("postgres://postgres:%s@localhost:%s/%s?sslmode=disable", password, resource.GetPort("5432/tcp"), database)

	c := &testPostgresContainer{
		databaseURL: databaseURL,
		pool:        pool,
		resource:    resource,
	}
	return c, nil
}

func (c *testPostgresContainer) Connect() (*sql.DB, error) {
	var db *sql.DB
	err := c.pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", c.databaseURL)
		if err != nil {
			return err
		}
		return db.Ping()
	})

	return db, err
}

func (c *testPostgresContainer) Close() error {
	return c.pool.Purge(c.resource)
}

func (s *ExampleTestSuite) TestExample() {
	s.Equal(5, 5)
}

func (s *ExampleTestSuite) TestExample2() {
	s.Equal(5, 5)
}

func (s *ExampleTestSuite) TestExample3() {
	s.NoError(s.db.Ping())
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, &ExampleTestSuite{})
}
