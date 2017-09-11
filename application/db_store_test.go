package application

import (
	"sync"
	"testing"
)

const getCreatedTablesSqliteSQL = "SELECT tbl_name FROM sqlite_master where type = 'table'"

func TestDBConnectionInitiation(t *testing.T) {
	loadConfiguration = GetTestConfiguration()
	locationInitializer = sync.Once{}
	if !IsDBConnectionAvailable() {
		t.Error("DB connection should have been available")
	}
}

func TestGetDB(t *testing.T) {
	rows, err := GetDB().Raw(getCreatedTablesSqliteSQL).Rows()
	if err != nil {
		t.Error("Could not run SQL against connection retrieved", err)
	}
	defer rows.Close()
}

func TestAutoMigration(t *testing.T) {
	rows, err := GetDB().Raw(getCreatedTablesSqliteSQL).Rows()
	if err != nil {
		t.Error("Could not run SQL against connection retrieved")
	}
	expectedTableNames := []string{"user_models"}
	expectedTableNameAssertions := make(map[string]bool)
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		found := false
		for _, name := range expectedTableNames {
			if name == tableName {
				found = true
			}
		}
		expectedTableNameAssertions[tableName] = found
	}
	for _, tableName := range expectedTableNames {
		if !expectedTableNameAssertions[tableName] {
			t.Error("Could not find table name", tableName)
		}
	}
	defer rows.Close()
}

func TestCloseDB(t *testing.T) {
	if !CloseDB() {
		t.Error("DB Close should have been successful")
	}
}
