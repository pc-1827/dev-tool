package localapp_test

import (
	localapp "local_app"
	"testing"
)

func TestConnection(t *testing.T) {
	t.Run("check if we connect to system store DB", func(t *testing.T) {
		db, err := localapp.Connect()
		if db != nil {
			defer db.Close()
		}

		if err != nil {
			t.Errorf("unable to connect to DB got error: %q", err.Error())
		}
	})
}

func TestRequestsTableSchema(t *testing.T) {
	db, err := localapp.Connect()
	if err != nil {
		db.Close()
		t.Errorf("unable to connect to DB got error:%q", err.Error())
	}
	defer db.Close()
	t.Run("check the schema of the DB", func(t *testing.T) {

		reqSchema := map[string]string{
			"id":           "integer",
			"request_data": "json",
			"request_time": "timestamp without time zone",
		}

		rows, _ := db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'requests'")
		defer rows.Close()
		for rows.Next() {
			var column_name string
			var data_type string
			if err = rows.Scan(&column_name, &data_type); err != nil {
				t.Error(err.Error())
			}

			if val, ok := reqSchema[column_name]; !ok || val != data_type {
				t.Errorf("the table does not contain the requried schema, got col:%q or type:%q, wanted schema %v",
					column_name, data_type, reqSchema)
			}
		}
	})
}
