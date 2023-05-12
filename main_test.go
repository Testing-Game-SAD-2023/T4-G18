package main

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	var err error

	const postgresAddr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	db, err = gorm.Open(postgres.Open(postgresAddr), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	code := m.Run()

	os.Exit(code)
}

func TestCleanup(t *testing.T) {
	t.SkipNow()

	seed(t, db)

	n, err := cleanup(db)

	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatalf("expected n=2; got n=%d", n)
	}

}

func seed(t *testing.T, db *gorm.DB) {
	t.Helper()

	tmpDir := os.TempDir()
	f1, err := os.CreateTemp(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer f1.Close()
	f1.Write([]byte("jfj"))

	f2, err := os.CreateTemp(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	f2.Write([]byte("jfj"))

	toInsert := []MetadataModel{
		{
			Path:   f1.Name(),
			TurnID: sql.NullInt64{Valid: false},
		},
		{
			Path:   f2.Name(),
			TurnID: sql.NullInt64{Valid: false},
		},
	}

	if err := db.Create(&toInsert).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := db.
			Session(&gorm.Session{AllowGlobalUpdate: true}).
			Delete(&MetadataModel{}).
			Error

		if err != nil {
			t.Fatal(err)
		}

		os.Remove(f1.Name())
		os.Remove(f2.Name())
	})
}
