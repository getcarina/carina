package makecoe

import (
	"net/http"
	"os"
	"testing"

	"fmt"
	"net/http/httptest"

	"github.com/getcarina/carina/common"
)

func TestDeleteClusterThenErrorOutHaltsWait(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.RequestURI {
		case "/":
			fmt.Fprintln(w, `{"versions": [{"status": "current", "min_version": "1.0", "max_version": "1.0", "id": "v1.0"}]}`)
		case "/clusters/99999999-9999-9999-9999-999999999999":
			fmt.Fprintln(w, `{ "status": "error" }`)
		default:
			fmt.Fprintln(w, "unexpected request: "+r.RequestURI)
			w.WriteHeader(404)
		}
	}))
	defer mockCarina.Close()

	acct := &Account{
		EndpointOverride:     mockCarina.URL,
		UserName:         os.Getenv("RS_USERNAME"),
		APIKey:           os.Getenv("RS_API_KEY"),
		Region:               "DFW",
	}

	svc := &MakeCOE{Account: acct}
	cluster := newCluster()
	cluster.Name = "test"
	cluster.ID = "99999999-9999-9999-9999-999999999999"

	err := svc.WaitUntilClusterIsDeleted(cluster)
	if err == nil {
		t.Error("WaitUntilClusterIsDeleted didn't stop when the cluster was in the error state.")
	} else {
		actualError := err.Error()
		expectedError := "Unable to delete cluster, an error occured while deleting."
		if expectedError != actualError {
			t.Errorf("Expected error: %s\nActual Error: %+v", expectedError, err)

		}
	}
}
