package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"code.sajari.com/docconv/v2"
	"github.com/lukasjarosch/go-docx"
	"github.com/marti700/templater/conf"
	"github.com/marti700/templater/customer"
)

type DocMetadata struct {
	Document string
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
		if len(attrs) == 1 {
			className := subMatch
			return fmt.Sprintf(`<img class='%s' type="image" hx-trigger="click" hx-target="#customer-selection" hx-get="/customer/selec?p=%s" data-bs-toggle="modal" data-bs-target="#customer-selection" src="https://upload.wikimedia.org/wikipedia/commons/0/0e/Add_user_icon_%%28blue%%29.svg" style="cursor: pointer; width: 2%%; height: 2%%"; ></img>`, className, className[len(className)-1:])
		} else {
			className := attrs[0]
			hiddenAttr := attrs[1]
			return fmt.Sprintf(`<img class='%s' %s type="image" hx-trigger="click" hx-target="#customer-selection" hx-get="/customer/select?p=%s" data-bs-toggle="modal" data-bs-target="#customer-selection" src="https://upload.wikimedia.org/wikipedia/commons/0/0e/Add_user_icon_%%28blue%%29.svg" style="cursor: pointer; width: 2%%; height: 2%%";></img>`, className, hiddenAttr, className[len(className)-1:])
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

func main() {
	dbConfig := conf.DBConfig{
		Host:     os.Getenv("POSTGRES_CUSTOMER_SERVER_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_CUSTOMER_SERVER_USER_NAME"),
		Password: os.Getenv("POSTGRES_CUSTOMER_SERVER_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_CUSTOMER_SERVER_DB_NAME"),
	}

	res, err := docconv.ConvertPath("Acto de Venta Alfredo Mateo4.docx")
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(res.Body)

	// inputText := template.Must(template.New("inputText").Parse(`<input type="text" value='' placehoder=''></input>`))

	additionalAttributes := `type="image" hx-trigger="click" hx-target="#customer-selection" hx-get="/customer/select" data-bs-toggle="modal" data-bs-target="#customer-selection" src="https://upload.wikimedia.org/wikipedia/commons/0/0e/Add_user_icon_%28blue%29.svg" style="cursor: pointer; width: 2%; height: 2%;"`

	metadata = DocMetadata{
		Document: replaceEmptyLines(replaceImgPlaceHolders(replaceDropdownPlaceholders(replaceInputPlaceholders(res.Body)), additionalAttributes)),
	}

	//get all customers
	http.HandleFunc("/customer/newCustomer", customer.CreateCustomer(dbConfig, "new-customer.html"))
	http.HandleFunc("/customer/updateCustomer", customer.UpdateCustomer(dbConfig, "edit-customer.html"))
	http.HandleFunc("/customers", customer.GetAllCustomers(dbConfig, "customers.html"))
	http.HandleFunc("/customer", customer.GetCustomerById(dbConfig, "customer.html"))
	http.HandleFunc("/customer/select", customer.SelectCustomer(dbConfig, "customer-selection.html"))
	// http.HandleFunc("/customer/customer", customer.UpdateCustomer(dbConfig))

	http.HandleFunc("/document", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("preview.html"))
		err := tmpl.Execute(w, metadata)
		if err != nil {
			log.Fatal(err.Error())
		}
		// w.Write([]byte(res.Body))
	})
	http.HandleFunc("/document/update", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Tamo aqui")
		cusID := r.URL.Query()["id"][0]
		c, _ := customer.FindCustomerById(dbConfig, cusID)
		fmt.Println(c)
		// tmpl := template.Must(template.ParseFiles("preview.html"))
		// err := tmpl.Execute(w, metadata)
		// if err != nil {
		// 	log.Fatal(err.Error())
		// }
		// w.Write([]byte(res.Body))
	})

	http.HandleFunc("/document/create", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			errMsg := []byte("Error parsing form")
			w.Write(errMsg)
		}
		f := r.Form
		file, err := docx.Open("Acto de Venta Alfredo Mateo.docx")
		if err != nil {
			log.Fatal(err.Error())
		}

		defer file.Close()

		placeholders := stringStringToIntfMap(f)

		file.ReplaceAll(placeholders)
		file.WriteToFile("substitution.docx")
	})

	fmt.Println("Executing server...")
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Server Running")

}
