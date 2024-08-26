package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/marti700/templater/conf"
	"github.com/marti700/templater/customer"
	"github.com/marti700/templater/document"
)

func main() {
	dbConfig := conf.DBConfig{
		Host:     os.Getenv("POSTGRES_CUSTOMER_SERVER_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_CUSTOMER_SERVER_USER_NAME"),
		Password: os.Getenv("POSTGRES_CUSTOMER_SERVER_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_CUSTOMER_SERVER_DB_NAME"),
	}

	// Customer routes
	http.HandleFunc("/customer/newCustomer", customer.CreateCustomer(dbConfig, "new-customer.html"))
	http.HandleFunc("/customer/updateCustomer", customer.UpdateCustomer(dbConfig, "edit-customer.html"))
	http.HandleFunc("/customers", customer.GetAllCustomers(dbConfig, "customers.html"))
	http.HandleFunc("/customer", customer.GetCustomerById(dbConfig, "customer.html"))
	http.HandleFunc("/customer/select", customer.SelectCustomer(dbConfig, "customer-selection.html"))

	// Document routes
	http.HandleFunc("/document", document.DocumentPreview("./tmpls/", "./sections.html"))
	http.HandleFunc("/document/create", document.CreteDocument("./tmpls/"))
	http.HandleFunc("/document/template/upload", document.Uploadtemplate("./tmpls/", "./templates.html"))
	http.HandleFunc("/document/templates", document.GetTemplatesList("./tmpls/", "./templates.html"))
	http.HandleFunc("/document/new", document.NewDocument("./tmpls/", "./document-selection.html"))

	http.HandleFunc("/wizard/customer/add", document.AddCustomer)

	// main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./home.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Executing server...")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Server Running")

}
