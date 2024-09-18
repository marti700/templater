package document

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentPreview(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:9090/document?template=Acto de Venta.docx", nil)
	if err != nil {
		log.Fatal(err)
	}
	//
	reqRecorder := responseRecorderHelper(req, DocumentPreview("./testTemplates/", "../preview.html"))

	// Check the status code and the body of the response.
	if status := reqRecorder.Code; status != http.StatusOK {
		t.Error("handler returned wrong status code:", status)
	}

	testResponse(t, reqRecorder, "./expectedResponses/getPreview.html")

}

func TestSectionsAdding(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:9090/document/sections/add", nil)

	if err != nil {
		log.Fatal(err)
	}
	//
	reqRecorder := responseRecorderHelper(req, NewSection("./testTemplates/"))

	// Check the status code and the body of the response.
	if status := reqRecorder.Code; status != http.StatusOK {
		t.Error("handler returned wrong status code:", status)
	}

	expectedResponse := `<div>
				<label>Nombre:</label> <input type="text" name="section-name"/> <label> Tipo: </label> <select name="section-type"> <option> Seleccion de cliente </option> <option> sub-plantilla</option></select>
			</div>`

	actualResponse := reqRecorder.Body.String()

	if expectedResponse != actualResponse {
		t.Error("Expected response is ", expectedResponse, "but was: ", actualResponse)
	}
}

func TestTemplateList(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:9090/document/templates", nil)
	if err != nil {
		log.Fatal(err)
	}
	//
	reqRecorder := responseRecorderHelper(req, GetTemplatesList("../tmpls", "../templates.html"))

	// Check the status code and the body of the response.
	if status := reqRecorder.Code; status != http.StatusOK {
		t.Error("handler returned wrong status code:", status)
	}

	testResponse(t, reqRecorder, "./expectedResponses/getAvailableTemplates.html")

}

func TestNewDocument(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:9090/document/templates", nil)
	if err != nil {
		log.Fatal(err)
	}
	//
	reqRecorder := responseRecorderHelper(req, GetTemplatesList("../tmpls", "../document-selection.html"))

	// Check the status code and the body of the response.
	if status := reqRecorder.Code; status != http.StatusOK {
		t.Error("handler returned wrong status code:", status)
	}

	testResponse(t, reqRecorder, "./expectedResponses/getDocumentSelectionModal.html")

}

func TestCreateDocument(t *testing.T) {

	form := url.Values{}
	form.Add("nombre_alcaldesa", "La alcandesa")
	form.Add("nombre_remitente", "el remitente")

	req, err := http.NewRequest("POST", "http://localhost:9090/document/create?template=carta_test.docx", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(CreteDocument("../tmpls/"))

	handler.ServeHTTP(reqRecorder, req)

	newDoc, err := os.Open("./substitution.docx")

	if err != nil {
		log.Fatal(err.Error())
	}

	defer newDoc.Close()
	defer os.Remove(newDoc.Name())

	data, err := io.ReadAll(newDoc)

	if err != nil {
		log.Fatal(err.Error())
	}

	if base64EncodedFile != base64.StdEncoding.EncodeToString(data) {
		t.Error("File was not generated correctly")
	}
}

func TestUploadTemplate(t *testing.T) {
	file, err := os.Open("./testFiles/testTemplate.txt")
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("template", filepath.Base(file.Name()))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(part, file); err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9094/document/template/upload", &buf)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	writer.Close()
	reqRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(Uploadtemplate("./testFiles/", "../templates.html"))

	handler.ServeHTTP(reqRecorder, req)

	// Check the status code and the body of the response.
	if status := reqRecorder.Code; status != http.StatusOK {
		fmt.Println("handler returned wrong status code:", status)
	}

	uploadedFile, err := os.Open("./testFiles/testTemplate/testTemplate.txt")

	if err != nil && uploadedFile.Name() != "testTemplate.txt" {
		t.Error("File was not correctly uploaded")
	}
	configFile, err := os.Open("./testFiles/testTemplate/cfg.txt")

	if err != nil && configFile.Name() != "cfg.txt" {
		t.Error("Section configuration file was not correctly saved")
	}

	err = os.RemoveAll("./testFiles/testTemplate/")

	if err != nil {
		fmt.Println(err.Error())
	}
}

func responseRecorderHelper(req *http.Request,
	handler func(http.ResponseWriter, *http.Request)) *httptest.ResponseRecorder {

	// ctx := context.Background()
	// container := DbContainer()
	// testDBhost, _ := container.ContainerIP(ctx)

	// dbConfig := SetUp(testDBhost)

	reqRecorder := httptest.NewRecorder()
	h := http.HandlerFunc(handler)
	h.ServeHTTP(reqRecorder, req)

	return reqRecorder
}

func testResponse(t *testing.T, reqRecorder *httptest.ResponseRecorder, expectedResponseFilePath string) {
	// Check the if the returned html is the correct one.
	f, err := os.ReadFile(expectedResponseFilePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	actualResponse := reqRecorder.Body.String()
	expectedResponse := string(f)
	os.WriteFile("tt.html", reqRecorder.Body.Bytes(), 0755)

	fmt.Println(actualResponse)
	if strings.Trim(actualResponse, "\n") != strings.Trim(expectedResponse, "\n") {
		t.Error("Returned HTML is not the correct one")
	}
}

const base64EncodedFile = "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"
