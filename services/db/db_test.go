package db

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"context"
	"database/sql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestInitDatabase(t *testing.T) {
	err := InitDatabase("mysqlfake", "fakesource")
	if assert.Error(t, err) {
		assert.Errorf(t, err, "sql: unknown driver \"mysqlfake\" (forgotten import?)")
	}
}

func TestBeginTxx(t *testing.T) {
	ctx := new(context.Context)
	opt := sql.TxOptions{sql.LevelDefault, true}
	_, err := BeginTxx(*ctx, &opt)
	if assert.Error(t, err) {
		assert.EqualError(t, ErrNotInitialized, err.Error())
	}
	_ = new(sqlmock.Sqlmock)
	err = InitDatabase("sqlmock", "test")
	/*ctx = new(context.Context)
	opt = sql.TxOptions{sql.LevelDefault, true}
	_, err = BeginTxx(*ctx, &opt)
	fmt.Println(err)*/
}