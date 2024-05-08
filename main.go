package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"code.sajari.com/docconv/v2"
)

type DocMetadata struct {
	Document string
}

func main() {
	res, err := docconv.ConvertPath("Acto de Venta Alfredo Mateo.docx")
	if err != nil {
		log.Fatal(err.Error())
	}
	metadata := DocMetadata{
		Document: res.Body,
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
