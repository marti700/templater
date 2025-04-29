package document

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	// "code.sajari.com/docconv/v2"
	"github.com/lukasjarosch/go-docx"
	"github.com/marti700/templater/conf"
	"github.com/marti700/templater/customer"
)

// //////
type DocMetadata struct {
	Document string
	Template string
}

type Document struct {
	Name     string
	Template string
}

func replaceInputPlaceholders(input string) string {
	re := regexp.MustCompile(`\{([^}]+):input\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		placeholder := re.FindStringSubmatch(match)
		return fmt.Sprintf("<input type=\"text\" name='%s' placeholder='%s'></input>",
			removecurlyBrackets(placeholder[0]), placeholder[1])
	})
}

func replaceEmptyLines(text string) string {
	// Define a regular expression that matches one or more empty lines
	pattern := `\n{2,}`
	re := regexp.MustCompile(pattern)

	// Replace function to insert a paragraph tag
	replace := func(match string) string {
		return fmt.Sprintf("</p>%s<p>", match)
	}

	// Substitute empty lines with the replace function
	return re.ReplaceAllStringFunc(text, replace)
}

func generateDropdown(name string, options []string) string {
	dropdown := fmt.Sprintf("<select name='%s'>\n", removecurlyBrackets(name))
	for _, option := range options {
		option = strings.TrimSpace(option)
		dropdown += fmt.Sprintf("<option>%s</option>\n", option)
	}
	dropdown += "</select>"
	return dropdown
}

func replaceDropdownPlaceholders(input string) string {
	re := regexp.MustCompile(`\{[^:]+:drop;([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the options part and split by ';'
		name := re.FindStringSubmatch(match)[0]
		optionsPart := re.FindStringSubmatch(match)[1]
		options := strings.Split(optionsPart, ";")
		// Generate the dropdown HTML for the options
		return generateDropdown(name, options)
	})
}

func replaceImgPlaceHolders(input string, additionalAttrs string) string {
	re := regexp.MustCompile(`\{img:([^}]*)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		subMatch := re.FindStringSubmatch(match)[1]
		attrs := strings.Split(subMatch, ";")
		imgSrc := make(map[string]string)
		imgSrc["add"] = "https://upload.wikimedia.org/wikipedia/commons/0/0e/Add_user_icon_%28blue%29.svg"
		imgSrc["delete"] = "https://icons.iconarchive.com/icons/visualpharm/must-have/256/Remove-icon.png"
		if len(attrs) == 1 {
			className := subMatch
			return fmt.Sprintf(`<img id='%s' type="image" hx-trigger="click" hx-target="#customer-selection" hx-get="/customer/select?p=%s" data-bs-toggle="modal" data-bs-target="#customer-selection" src="%s" style="cursor: pointer; width: 2%%; height: 2%%"; ></img>`, className, className[len(className)-1:], imgSrc[className[:len(className)-1]])
		} else {
			className := attrs[0]
			hiddenAttr := attrs[1]
			return fmt.Sprintf(`<img id='%s' %s type="image" src="%s" onClick=clearCustomer(%s) style="cursor: pointer; width: 2%%; height: 2%%";></img>`, className, hiddenAttr, imgSrc[className[:len(className)-1]], className[len(className)-1:])
		}
	})
}

func removecurlyBrackets(name string) string {
	nName := strings.ReplaceAll(name, "{", "")
	nName = strings.ReplaceAll(nName, "}", "")

	return nName
}

func stringStringToIntfMap(strMap map[string][]string) map[string]interface{} {
	intfMap := make(map[string]interface{}, len(strMap))
	for key, value := range strMap {
		// if a key have multiple values we just one the first one, a document placeholder can't have multiple values
		intfMap[key] = value[0]
	}
	return intfMap
}

// func DocumentPreview(templatesFolderPath string, viewTemplatesPath string) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		templateName := r.URL.Query()["template"][0]
// 		res, err := docconv.ConvertPath(templatesFolderPath + templateName)
// 		if err != nil {
// 			// log.Fatal(err.Error())
// 			fmt.Println(err.Error())
// 		}
// 		// fmt.Println(res.Body)

// 		additionalAttributes := `type="image" hx-trigger="click" hx-target="#customer-selection" hx-get="/customer/select" data-bs-toggle="modal" data-bs-target="#customer-selection" src="https://upload.wikimedia.org/wikipedia/commons/0/0e/Add_user_icon_%28blue%29.svg" style="cursor: pointer; width: 2%; height: 2%;"`
// 		metadata := DocMetadata{
// 			Document: replaceEmptyLines(replaceImgPlaceHolders(replaceDropdownPlaceholders(replaceInputPlaceholders(res.Body)), additionalAttributes)),
// 			Template: templateName,
// 		}

// 		tmpl := template.Must(template.ParseFiles(viewTemplatesPath))
// 		err = tmpl.Execute(w, metadata)
// 		if err != nil {
// 			log.Fatal(err.Error())
// 		}
// 	}
// }

func DocumentPreview(dbconf conf.DBConfig, viewTemplatesPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		docName := r.URL.Query()["name"][0]

		doc, err := findDocumentsByName(dbconf, docName)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles(viewTemplatesPath))
		err = tmpl.Execute(w, doc)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func findDocumentsByName(dbconf conf.DBConfig, documentName string) (Document, error) {

	db := dbconf.DbConn()
	var d Document
	stmt, err := db.Prepare("Select name, template from documents where name = $1")
	if err != nil {
		return Document{}, err

	}

	defer stmt.Close()
	stmt.QueryRow(documentName).Scan(&d.Name, &d.Template)

	return d, nil
}

func GenerateDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	document := r.FormValue("html")

	// Get the URL from the env variable or default to localhost:5000
	url := os.Getenv("PYTHON_SERVICE_URL")
	if url == "" {
		url = "http://localhost:5000/generate_and_save" // Or /generate_and_save
	}

	// Create a new POST request with the HTML content as the body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(document)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", resp.Status)
		return
	}

	// Decode the JSON response
	var result map[string]string // Use a map to handle the JSON

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// the service returns the keys document and error if the generate_and_encode route was invoked
	// and the keys message and error if the generate_and_save method was invoked
	if val, ok := result["document"]; ok { // Check if the "document" key exists
		fmt.Println("Base64 encoded document:", val)
		// TODO: decode the base64 string in Go)
	} else if val, ok := result["message"]; ok { // Check for the "message" key
		fmt.Println("Message:", val)
		// TODO: (Handle the message from the /generate_and_save endpoint)
	} else if val, ok := result["error"]; ok {
		fmt.Println("Error from service:", val)
	}

	fmt.Println(document)
}

func CreteDocument(templateFolderPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		templateName := r.URL.Query()["template"][0]
		err := r.ParseForm()
		if err != nil {
			errMsg := []byte("Error parsing form")
			w.Write(errMsg)
		}
		f := r.Form
		file, err := docx.Open(templateFolderPath + templateName)
		if err != nil {
			log.Fatal(err.Error())
		}

		defer file.Close()

		placeholders := stringStringToIntfMap(f)

		file.ReplaceAll(placeholders)
		file.WriteToFile("substitution.docx")
	}
}

func templateNames(templatesPath string) ([]string, error) {
	fls, err := os.ReadDir(templatesPath)
	if err != nil {
		return nil, err
	}
	fileNames := make([]string, len(fls))
	for i, fn := range fls {
		fileNames[i] = fn.Name()
	}

	return fileNames, nil
}

func Uploadtemplate(templatesFolderPath, templatePath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// err := r.ParseMultipartForm(10 << 20)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return

		// }
		// documentFile, header, err := r.FormFile("template")
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// defer documentFile.Close()
		// fileName := header.Filename
		// folderName := fileName[:strings.LastIndex(fileName, ".")]
		// err = os.MkdirAll(templatesFolderPath+folderName, os.ModePerm)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// filePath := templatesFolderPath + folderName + "/" + fileName
		// dst, err := os.Create(filePath)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// defer dst.Close()

		// if _, err := io.Copy(dst, documentFile); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// saveDocSections(r.Form["section-name"], r.Form["section-type"], templatesFolderPath+folderName+"/")

		// fileNames, err := templateNames(templatesFolderPath)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		w.Header().Set("Content-Type", "text/html")
		tmpl := template.Must(template.ParseFiles(templatePath))
		// err = tmpl.Execute(w, fileNames)
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func SaveTemplate(dbconf conf.DBConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		db := dbconf.DbConn()
		stmt, err := db.Prepare("INSERT INTO documents (name, template) VALUES ($1, $2)")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer stmt.Close()

		res, err := stmt.Exec(r.FormValue("templateName"), r.FormValue("html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirectURL := url.URL{Scheme: "http", Host: r.Host, Path: "/document/templates"}
		// w.Header().Set("Location", redirectURL.String())
		// w.WriteHeader(http.StatusSeeOther)

		http.Redirect(w, r, "/document/templates", http.StatusSeeOther)
	}
}

func saveDocSections(names, types []string, path string) error {
	file, err := os.Create(path + "cfg.txt")

	if err != nil {
		return err
	}
	defer file.Close()

	// Create a buffered writer for efficient writing
	writer := bufio.NewWriter(file)

	// Iterate over both slices simultaneously
	for i := 0; i < len(names); i++ {
		line := fmt.Sprintf("%s;%s\n", names[i], types[i])
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// Flush the buffer to ensure data is written
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func GetTemplatesList(dbconf conf.DBConfig, templatePath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// fileNames, err := templateNames(templatesFolderPath)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		templates, err := findAllTemplates(dbconf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles(templatePath))
		err = tmpl.Execute(w, templates)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func findAllTemplates(dbconf conf.DBConfig) ([]Document, error) {
	db := dbconf.DbConn()
	rows, err := db.Query("Select name, template from documents")
	if err != nil {
		return nil, err
	}

	var documents []Document

	for rows.Next() {
		doc := Document{}
		err := rows.Scan(&doc.Name, &doc.Template)
		if err != nil {
			log.Fatal(err.Error())
		}
		documents = append(documents, doc)
	}

	return documents, nil

}

func NewDocument(templatesFolderPath, templatePath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fileNames, err := templateNames(templatesFolderPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("./document-selection.html"))
		err = tmpl.Execute(w, fileNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func NewSection(templatesFolderPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := template.Must(template.New("doc-sections").Parse(
			`<div>
				<label>Nombre:</label> <input type="text" name="section-name"/> <label> Tipo: </label> <select name="section-type"> <option> Seleccion de cliente </option> <option> sub-plantilla</option></select>
			</div>`)).Execute(w, nil)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	// params := make(map[string]string)

	cus := customer.NewCustomerEntity(
		query["id"][0],
		query["idtype"][0],
		query["name"][0],
		query["lastname"][0],
		query["address"][0],
		query["nationality"][0],
		query["ocupation"][0],
		query["civilstatus"][0],
		query["gender"][0],
	)

	// for key, values := range query {
	// 	params[key] = values[0] // Assuming you only need the first value
	// }

	tmpl := template.Must(template.New("wizard-customer-select").Parse(
		`<div id={{.ID}}>
            <label>{{.Name}}</label>
			<div hidden>
				<input type="text" name="id" value="{{.ID}}"><br />
				<input type="text" name="idType" value="{{.ID}}"><br />
				<input type="text" name="name" value="{{.Name}}"><br />
				<input type="text" name="lastname" value="{{.LastName}}"><br />
				<input type="text" name="ocupation" value="{{.Ocupation}}"><br />
				<input type="text" name="nationality" value="{{.Nationality}}"><br />
				<input type="text" name="civilStatus" value="{{.CivilStatus}}"><br />
				<input type="text" name="gender" value="{{.Gender}}">
				<input type="text" name="address" value="{{.Address}}">
			</div>
		</div>`))
	err := tmpl.Execute(w, cus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
