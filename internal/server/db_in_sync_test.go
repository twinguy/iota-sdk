package server

import (
	"github.com/iota-agency/iota-sdk/internal/testutils"
	"github.com/iota-agency/iota-sdk/pkg/dbutils"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestCheckModels(t *testing.T) { //nolint:paralleltest
	ctx := testutils.GetTestContext()
	if err := dbutils.CheckModels(ctx.GormDB, RegisteredModels); err != nil {
		t.Fatal(err)
	}
}
