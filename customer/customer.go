package customer

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

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
func CreateCustomer(dbconf conf.DBConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := dbconf.DbConn()

		c := Customer{
			ID:          "1",
			Name:        "name",
			LastName:    "lastName",
			Address:     "address",
			Nationality: "nationality",
			Ocupation:   "ocupation",
			CivilStatus: "civilStatus",
			Gender:      "gender",
		}
		stmt, err := db.Prepare("INSERT INTO customers VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer stmt.Close()

		res, err := stmt.Exec(c.ID, c.IDType, c.Name, c.LastName, c.Address, c.Nationality, c.Ocupation, c.CivilStatus, c.Gender)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func UpdateCustomer(dbconf conf.DBConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := dbconf.DbConn()
		stmt, err := db.Prepare("UPDATE customers SET name = $1, last_name = $2, address = $3, nationality = $4, occupation =  $5, civil_status = $6 where id = $7")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer stmt.Close()

		c := Customer{
			ID:          "1",
			IDType:      "IDType",
			Name:        "name13",
			LastName:    "lastName",
			Address:     "address",
			Nationality: "nationality",
			Ocupation:   "ocupation",
			CivilStatus: "civilStatus",
			Gender:      "gender",
		}

		res, err := stmt.Exec(c.Name,
			c.LastName,
			c.Address,
			c.Nationality,
			c.Ocupation,
			c.CivilStatus,
			c.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetAllCustomers(dbconf conf.DBConfig, templatePath string) func(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusOK)
		parseTemplate(customers, templatePath, w)
		fmt.Printf("All OK")
	}
}

func GetCustomerById(dbconf conf.DBConfig, templatePath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		customerId := strings.TrimSpace(r.URL.Query()["id"][0])

		db := dbconf.DbConn()
		var c Customer
		stmt, err := db.Prepare("Select * from customers where id = $1")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer stmt.Close()
		stmt.QueryRow(customerId).Scan(&c.ID, &c.IDType, &c.Name, &c.LastName, &c.Address, &c.Nationality, &c.Ocupation, &c.CivilStatus, &c.Gender)

		w.WriteHeader(http.StatusOK)
		parseTemplate(c, templatePath, w)
	}
}

func parseTemplate(obj any, templatePath string, writter io.Writer) {
	tmpl := template.Must(template.ParseFiles(templatePath))
	err := tmpl.Execute(writter, obj)
	if err != nil {
		log.Fatal(err)
	}
}
