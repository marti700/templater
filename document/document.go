package document

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"code.sajari.com/docconv/v2"
	"github.com/lukasjarosch/go-docx"
)

// //////
type DocMetadata struct {
	Document string
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

func DocumentPreview(templatesFolderPath string, viewTemplatesPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		templateName := r.URL.Query()["template"][0]
		res, err := docconv.ConvertPath(templatesFolderPath + templateName)
		if err != nil {
			// log.Fatal(err.Error())
			fmt.Println(err.Error())
		}
		// fmt.Println(res.Body)

		additionalAttributes := `type="image" hx-trigger="click" hx-target="#customer-selection" hx-get="/customer/select" data-bs-toggle="modal" data-bs-target="#customer-selection" src="https://upload.wikimedia.org/wikipedia/commons/0/0e/Add_user_icon_%28blue%29.svg" style="cursor: pointer; width: 2%; height: 2%;"`
		metadata := DocMetadata{
			Document: replaceEmptyLines(replaceImgPlaceHolders(replaceDropdownPlaceholders(replaceInputPlaceholders(res.Body)), additionalAttributes)),
			Template: templateName,
		}

		tmpl := template.Must(template.ParseFiles(viewTemplatesPath))
		err = tmpl.Execute(w, metadata)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
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
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}
		documentFile, header, err := r.FormFile("template")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer documentFile.Close()
		fileName := header.Filename
		filePath := templatesFolderPath + fileName

		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, documentFile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileNames, err := templateNames(templatesFolderPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl := template.Must(template.ParseFiles(templatePath))
		err = tmpl.Execute(w, fileNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetTemplatesList(templatesFolderPath, templatePath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fileNames, err := templateNames(templatesFolderPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles(templatePath))
		err = tmpl.Execute(w, fileNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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
