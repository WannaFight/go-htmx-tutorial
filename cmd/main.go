package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

var id = 0

type Contact struct {
	Name  string
	Email string
	Id    int
}

func newContact(name, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
	}
}

type Contacts []Contact

// Find contact by specific strategy and return its index, return -1 otherwise
func (c Contacts) find(strategy func(c Contact) bool) int {
	for i, contact := range c {
		if strategy(contact) {
			return i
		}
	}
	return -1
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Data struct {
	Contacts Contacts
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("jon", "jon@mail.ru"),
			newContact("bob", "bob@mail.ru"),
			newContact("duke", "duke@mail.ru"),
		},
	}
}

type Page struct {
	Form FormData
	Data Data
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	page := newPage()
	e.Static("/images", "images")
	e.Static("/css", "css")
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.Contacts.find(func(c Contact) bool { return c.Email == email }) > -1 {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"

			return c.Render(http.StatusBadRequest, "form", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		c.Render(http.StatusOK, "form", newFormData())
		return c.Render(http.StatusOK, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid id")
		}

		index := page.Data.Contacts.find(func(c Contact) bool { return c.Id == id })
		if index == -1 {
			return c.String(http.StatusBadRequest, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(http.StatusNoContent)

	})

	e.Logger.Fatal(e.Start(":8080"))
}
