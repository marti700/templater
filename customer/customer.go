package customer

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/marti700/templater/conf"
)

type Customer struct {
	ID          string
	IDType      string
	Name        string
	LastName    string
	Address     string
	Nationality string
	Ocupation   string
	CivilStatus string
	Gender      string
}

// A constructor function used to create a customer
func NewCustomerEntity(ID, IDType, name, lastName, address, nationality, ocupation, civliStatus, gender string) Customer {

	cust := Customer{
		ID:          ID,
		IDType:      IDType,
		Name:        name,
		LastName:    lastName,
		Address:     address,
		Nationality: nationality,
		Ocupation:   ocupation,
		CivilStatus: civliStatus,
		Gender:      gender,
	}

	return cust
}

func GetAllCustomers(dbconf conf.DBConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := dbconf.DbConn()

		rows, err := db.Query("Select * from customers")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var customers []Customer
		for rows.Next() {
			cus := Customer{}
			err := rows.Scan(&cus.ID, &cus.IDType, &cus.Name, &cus.LastName, &cus.Address, &cus.Nationality, &cus.Ocupation, &cus.CivilStatus, &cus.Gender)
			if err != nil {
				log.Fatal(err.Error())
			}
			customers = append(customers, cus)
		}

		// parseTemplate(customers, "path to template", w)
		fmt.Printf("All OK")
	}
}

func parseTemplate(obj any, templatePath string, writter io.Writer) {
	tmpl := template.Must(template.ParseFiles(templatePath))
	err := tmpl.Execute(writter, obj)
	if err != nil {
		log.Fatal(err)
	}
}
