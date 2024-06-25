package customer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/marti700/templater/conf"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetUp(host string) conf.DBConfig {
	return conf.DBConfig{
		Host:     host,
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "testdb",
	}
}

func DbContainer() *postgres.PostgresContainer {
	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		postgres.WithInitScripts(filepath.Join("./", "init-db.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatal(err)
	}
	// nets := []string {"test-net"}
	nets, err := pgContainer.Networks(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(nets)

	return pgContainer
}

func TestSaveCustomer(t *testing.T) {

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("id", "1234")
	w.WriteField("idType", "passport")
	w.WriteField("name", "Test")
	w.WriteField("lastname", "The Tester")
	w.WriteField("address", "Test Avenue #87")
	w.WriteField("nationality", "Testlandian")
	w.WriteField("ocupation", "Tester")
	w.WriteField("civilStatus", "SINGLE")

	w.Close()

	req, err := http.NewRequest("POST", "http://localhost:9090/newCustomer", &b)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	reqRecorder := customerTestUtil(req, "new-customer.html", CreateCustomer)

	// Check the status code and the body of the response.
	if status := reqRecorder.Code; status != http.StatusOK {
		t.Error("handler returned wrong status code:", status)
	}

}

func customerTestUtil(req *http.Request, templatePath string,
	handler func(conf.DBConfig, string) func(http.ResponseWriter, *http.Request)) *httptest.ResponseRecorder {

	ctx := context.Background()
	container := DbContainer()
	testDBhost, _ := container.ContainerIP(ctx)

	dbConfig := SetUp(testDBhost)

	reqRecorder := httptest.NewRecorder()
	h := http.HandlerFunc(handler(dbConfig, templatePath))
	h.ServeHTTP(reqRecorder, req)

	return reqRecorder
}
