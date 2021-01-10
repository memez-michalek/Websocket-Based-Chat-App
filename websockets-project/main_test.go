package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"syreclabs.com/go/faker"
)

var (
	formUsername = faker.Internet().UserName()
	formEmail    = faker.Internet().Email()
	formPassword = faker.Internet().Password(8, 14)
)

func TestMainHealthTest(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		log.Fatal(err)

	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(indexHandler)
	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong http status code. GOT: %v , WANT: %v", status, http.StatusOK)
	}

}

func TestLoginHealthTest(t *testing.T) {

	formdata := url.Values{
		"username": {formUsername},
		"email":    {formEmail},
		"password": {formPassword},
	}

	req, err := http.PostForm("http://localhost:8081/login/", formdata)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 200, req.StatusCode, "variables should have exact same values")

}
func TestRegister(t *testing.T) {
	registerDATA := url.Values{
		"username": {formUsername},
		"email":    {formEmail},
		"password": {formPassword},
	}
	req, err := http.PostForm("http://localhost:8081/login", registerDATA)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 200, req.StatusCode, "values should equal 200")

}
