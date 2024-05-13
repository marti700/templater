package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"text/template"

	"code.sajari.com/docconv/v2"
)

type DocMetadata struct {
	Document string
}

func replacePlaceholders(str string) string {
	// Define a regular expression to match placeholders
	pattern := `{([^{}]+):([^}]+)}`
	re := regexp.MustCompile(pattern)

	// Replace function to create the HTML input element
	replace := func(match string) string {
		placehoder := re.FindStringSubmatch(match)[1:] // Capture name and value from matched group
		return fmt.Sprintf("<input type=\"text\" name='%s' placeholder='%s'></input>", placehoder[0], placehoder[0])
	}

	// Substitute all placeholders with the replace function
	return re.ReplaceAllStringFunc(str, replace)
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

func main() {
	res, err := docconv.ConvertPath("Acto de Venta Alfredo Mateo.docx")
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(res.Body)

	// inputText := template.Must(template.New("inputText").Parse(`<input type="text" value='' placehoder=''></input>`))
	metadata := DocMetadata{
		Document: replaceEmptyLines(replacePlaceholders(res.Body)),
	}

	http.HandleFunc("/document", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("preview.html"))
		err := tmpl.Execute(w, metadata)
		if err != nil {
			log.Fatal(err.Error())
		}
		// w.Write([]byte(res.Body))
	})

	fmt.Println("Executing server...")
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Server Running")

}
