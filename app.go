package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("%s:%s@/%s?parseTime=true", user, password, dbname)

	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/account", a.getAccounts).Methods("GET")
	a.Router.HandleFunc("/account/{accountid:[0-9]+}", a.getAccount).Methods("GET")
	a.Router.HandleFunc("/account", a.createAccount).Methods("POST")

	a.Router.HandleFunc("/transaction", a.getTransactions).Methods("GET")
	a.Router.HandleFunc("/transaction", a.createTransaction).Methods("POST")
	a.Router.HandleFunc("/account/{accountid:[0-9]+}/history", a.getHistory).Methods("GET")
}

func (a *App) getAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountid, err := strconv.Atoi(vars["accountid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	acc := Account{AccountID: uint32(accountid)}

	err = acc.getAccount(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Account not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, acc)
}

func (a *App) getAccounts(w http.ResponseWriter, r *http.Request) {
	acc := Account{}

	accounts, err := acc.getAccounts(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Account not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, accounts)
}

func (a *App) getHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	accountid, err := strconv.Atoi(vars["accountid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	t := Transaction{DepositDest: uint32(accountid)}

	trans, err := t.getHistory(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "No transaction history found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, trans)

}

func (a *App) getTransactions(w http.ResponseWriter, r *http.Request) {
	t := Transaction{}

	trans, err := t.getTransactions(a.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "No transaction history found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, trans)

}

func (a *App) createAccount(w http.ResponseWriter, r *http.Request) {
	var acc Account
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&acc)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	if !isValidIDCard(acc.IDCard) {
		respondWithError(w, http.StatusBadRequest, "Invalid ID Card")
		return
	}

	if !isValidName(acc.Name) {
		respondWithError(w, http.StatusBadRequest, "Invalid Name")
		return
	}

	if !isValidEmail(acc.Email) {
		respondWithError(w, http.StatusBadRequest, "Invalid Email Address")
		return
	}

	err = acc.createAccount(a.DB)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			respondWithError(w, http.StatusBadRequest, "Account exists!")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, acc)
}

func (a *App) createTransaction(w http.ResponseWriter, r *http.Request) {
	var trans Transaction
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&trans)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	if trans.ExternalSource != "" && !isValidEmail(trans.ExternalSource) {
		respondWithError(w, http.StatusBadRequest, "Invalid Email Address")
		return
	}

	err = trans.createTransaction(a.DB)
	if err != nil {
		if strings.Contains(err.Error(), "a foreign key constraint fails") {
			respondWithError(w, http.StatusNotFound, "Invalid Account ID")
			return
		}

		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = sendDepositNotificationEmail(trans)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, trans)
}

func sendDepositNotificationEmail(trans Transaction) error {
	smtpserver := os.Getenv("SMTP_SERVER")
	smtpport := os.Getenv("SMTP_PORT")

	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	var to string
	if trans.InternalSourceEmail == "" {
		to = trans.ExternalSource
	} else {
		to = trans.InternalSourceEmail
	}

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Cash Deposit Notification\n\n" +
		"You have successfully deposited " + strconv.FormatFloat(trans.Amount, 'f', 2, 64) + " to account number " + strconv.Itoa(int(trans.DepositDest)) + ". Ref No. #" + strconv.Itoa(int(trans.TransactionID))

	err := smtp.SendMail(smtpserver+":"+smtpport, smtp.PlainAuth("", from, pass, smtpserver), from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

var isValidName = regexp.MustCompile(`^[a-zA-Z\s]{3,100}$`).MatchString
var isValidIDCard = regexp.MustCompile(`^[0-9]{16}$`).MatchString
var isValidEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$").MatchString
