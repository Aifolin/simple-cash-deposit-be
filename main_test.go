package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	. "github.com/mikeadityas/simple-cash-deposit-be"
)

var a App

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	a = App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	ensureTableExists()
	rand.Seed(time.Now().UnixNano())
	code := m.Run()

	clearTable()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/account", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentAccount(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/account/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Account not found" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Account not found'. Got '%s'", m["error"])
	}
}

func TestGetAccountAndBalance(t *testing.T) {
	clearTable()
	addAccount(1)

	req, _ := http.NewRequest("GET", "/account/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var acc Account
	acc = Account{}
	json.Unmarshal(response.Body.Bytes(), &acc)
	if acc.Balance != float64(0) {
		t.Errorf("Expected the 'balance' to be set to 0.00. Got '%.2f'", acc.Balance)
	}

	addTransaction(false, "michaeladityas@gmail.com", 1)

	req, _ = http.NewRequest("GET", "/account/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &acc)
	if acc.Balance != float64(100000) {
		t.Errorf("Expected the 'balance' to be set to 100000.00. Got '%.2f'", acc.Balance)
	}
}

func TestCreateDuplicateAccount(t *testing.T) {
	clearTable()

	idCard := generateString(48, 57, 16)
	name := generateString(97, 122, 7)
	email := name + "@mail.com"

	payload := []byte(`{"idcardno":"` + idCard + `","name":"` + name + `","email":"` + email + `"}`)

	req, _ := http.NewRequest("POST", "/account", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["idcardno"] != idCard {
		t.Errorf("Expected ID Card Number to be "+idCard+". Got %v", m["idcardno"])
	}

	if m["name"] != name {
		t.Errorf("Expected name to be "+name+". Got %v", m["idcardno"])
	}

	if m["email"] != email {
		t.Errorf("Expected email to be "+email+". Got %v", m["email"])
	}

	req, _ = http.NewRequest("POST", "/account", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Account exists!" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Account exists!'. Got '%s'", m["error"])
	}
}

func TestCreateAccountInvalidIDCardNumber(t *testing.T) {
	clearTable()

	idCard := "1234567890ABCDEF"
	name := generateString(97, 122, 7)
	email := name + "@mail.com"

	payload := []byte(`{"idcardno":"` + idCard + `","name":"` + name + `","email":"` + email + `"}`)

	req, _ := http.NewRequest("POST", "/account", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid ID Card" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid ID Card'. Got '%s'", m["error"])
	}
}

func TestCreateAccountInvalidName(t *testing.T) {
	clearTable()

	idCard := generateString(48, 57, 16)
	name := "th3nun"
	email := name + "@mail.com"

	payload := []byte(`{"idcardno":"` + idCard + `","name":"` + name + `","email":"` + email + `"}`)

	req, _ := http.NewRequest("POST", "/account", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid Name" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid Name'. Got '%s'", m["error"])
	}
}

func TestCreateAccountInvalidEmail(t *testing.T) {
	clearTable()

	idCard := generateString(48, 57, 16)
	name := generateString(97, 122, 7)
	email := name + "mail.com"

	payload := []byte(`{"idcardno":"` + idCard + `","name":"` + name + `","email":"` + email + `"}`)

	req, _ := http.NewRequest("POST", "/account", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid Email Address" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid Email Address'. Got '%s'", m["error"])
	}
}

func TestCreateDepositFromRegisteredAccount(t *testing.T) {
	clearTable()
	addAccount(1)

	// Create account with valid Email, so we can also test email functionality
	q := `INSERT INTO account
				(id_card_number, name, email)
				VALUES
				("1234567890123456","Michael","mike.sutiono@gmail.com")
		`
	a.DB.Exec(q)

	payload := []byte(`{"depositdest":1,"internalsource":2,"amount":3879000}`)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["depositdest"] != 1.0 {
		t.Errorf("Expected depositdest to be 1. Got %v", m["depositdest"])
	}

	if m["internalsource"] != 2.0 {
		t.Errorf("Expected internalsource to be 2. Got %v", m["internalsource"])
	}

	if m["amount"] != 3879000.0 {
		t.Errorf("Expected amount to be 3879000. Got %v", m["amount"])
	}

}

func TestCreateDepositFromUnregisteredAccount(t *testing.T) {
	clearTable()
	addAccount(1)

	payload := []byte(`{"depositdest":1,"externalsource":"michaeladityas@live.com","amount":3879000}`)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["depositdest"] != 1.0 {
		t.Errorf("Expected depositdest to be 1. Got %v", m["depositdest"])
	}

	if m["externalsource"] != "michaeladityas@live.com" {
		t.Errorf("Expected internalsource to be michaeladityas@live.com. Got %v", m["externalsource"])
	}

	if m["amount"] != 3879000.0 {
		t.Errorf("Expected amount to be 3879000. Got %v", m["amount"])
	}

}

func TestCreateDepositInvalidExternal(t *testing.T) {
	clearTable()
	addAccount(1)

	payload := []byte(`{"depositdest":1,"externalsource":"asd.com","amount":3879000}`)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid Email Address" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid Email Address'. Got '%s'", m["error"])
	}
}

func TestCreateDepositInvalidInternal(t *testing.T) {
	clearTable()
	addAccount(1)

	payload := []byte(`{"depositdest":1,"internalsource":2,"amount":3879000}`)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid Account ID" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid Account ID'. Got '%s'", m["error"])
	}
}

func TestCreateDepositInvalidDestination(t *testing.T) {
	clearTable()
	addAccount(1)

	payload := []byte(`{"depositdest":2,"internalsource":1,"amount":3879000}`)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid Account ID" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid Account ID'. Got '%s'", m["error"])
	}
}

func TestCreateDepositInvalidAmount(t *testing.T) {
	clearTable()
	addAccount(1)

	payload := []byte(`{"depositdest":1,"internalsource":1,"amount":"3879000aaa"}`)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid request payload" {
		t.Errorf("Expected the 'error' message of the response to be set to 'Invalid request payload'. Got '%s'", m["error"])
	}
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableAccountCreationQuery); err != nil {
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(tableTransactionLogCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM account")
	a.DB.Exec("DELETE FROM transaction_log")
	a.DB.Exec("ALTER TABLE account AUTO_INCREMENT = 1")
	a.DB.Exec("ALTER TABLE transaction_log AUTO_INCREMENT = 1")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addAccount(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		q := `INSERT INTO account
					(id_card_number, name, email)
					VALUES
					(?,?,?)
		`
		idCard := generateString(48, 57, 16)
		name := generateString(97, 122, 7)
		email := name + "@mail.com"
		a.DB.Exec(q, idCard, name, email)
	}
}

func addTransaction(isMember bool, from string, to int) {
	var q string
	if isMember {
		q = `INSERT INTO transaction_log
				 (source_internal, destination, amount)
				 VALUES
				 (?,?,?)`
	} else {
		q = `INSERT INTO transaction_log
				 (source_external, destination, amount)
				 VALUES
				 (?,?,?)`
	}
	a.DB.Exec(q, from, to, 100000)
}

func generateString(from int, to int, len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(from, to))
	}
	return string(bytes)
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

const tableAccountCreationQuery = `
CREATE TABLE IF NOT EXISTS account (
  account_id INT UNSIGNED NOT NULL AUTO_INCREMENT ,
  id_card_number VARCHAR(20) NOT NULL UNIQUE,
  name VARCHAR(100) NOT NULL ,
  email VARCHAR(320) NOT NULL ,
  registration_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
  PRIMARY KEY (account_id,id_card_number)
) ENGINE = InnoDB;
`

const tableTransactionLogCreationQuery = `
CREATE TABLE IF NOT EXISTS transaction_log (
  transaction_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT ,
  source_external VARCHAR(320) NULL ,
  source_internal INT UNSIGNED NULL ,
  destination INT UNSIGNED NOT NULL ,
  amount DOUBLE UNSIGNED NOT NULL ,
  transaction_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
  PRIMARY KEY (transaction_id),
  CONSTRAINT internal_account_destination FOREIGN KEY (destination) REFERENCES account (account_id) ON UPDATE CASCADE,
  CONSTRAINT internal_account_source FOREIGN KEY (source_internal) REFERENCES account (account_id) ON UPDATE CASCADE
) ENGINE = InnoDB;
`
