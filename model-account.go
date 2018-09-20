package main

import (
	"database/sql"
)

type Account struct {
	AccountID uint32  `json:"accountid"`
	IDCard    string  `json:"idcardno"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Balance   float64 `json:"balance"`
}

func (acc *Account) getAccount(db *sql.DB) error {
	var q string = `SELECT acc.id_card_number, acc.name, acc.email, COALESCE(SUM(tl.amount),0) AS balance
									FROM account acc
									LEFT JOIN transaction_log tl ON tl.destination = acc.account_id
									WHERE acc.account_id = ?
									HAVING acc.id_card_number IS NOT NULL`
	return db.QueryRow(q, acc.AccountID).Scan(&acc.IDCard, &acc.Name, &acc.Email, &acc.Balance)
}

func (acc *Account) createAccount(db *sql.DB) error {
	var q string = `INSERT INTO account
									(id_card_number, name, email)
									VALUES
									(?,?,?)`
	res, err := db.Exec(q, acc.IDCard, acc.Name, acc.Email)

	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	acc.AccountID = uint32(id)
	return nil
}

func (acc *Account) getAccounts(db *sql.DB) ([]Account, error) {
	var q string = `SELECT acc.account_id, acc.id_card_number, acc.name, acc.email, COALESCE(SUM(tl.amount),0) AS balance
									FROM account acc
									LEFT JOIN transaction_log tl ON tl.destination = acc.account_id
									GROUP BY acc.id_card_number
									HAVING acc.id_card_number IS NOT NULL
									ORDER BY tl.transaction_time DESC
									`
	rows, err := db.Query(q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var accnt Account
		err := rows.Scan(&accnt.AccountID, &accnt.IDCard, &accnt.Name, &accnt.Email, &accnt.Balance)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, accnt)
	}

	return accounts, nil
}
