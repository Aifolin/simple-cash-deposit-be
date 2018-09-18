package main

import (
	"database/sql"
	"time"
)

type Transaction struct {
	TransactionID       uint64    `json:"transid"`
	DepositDest         uint32    `json:"depositdest"`
	ExternalSource      string    `json:"externalsource"`
	InternalSource      uint32    `json:"internalsource"`
	InternalSourceEmail string    `json:"internalsourceemail"`
	Name                string    `json:"name"`
	Amount              float64   `json:"amount"`
	TransactionTime     time.Time `json:"transtime"`
}

func (trans *Transaction) getHistory(db *sql.DB) ([]Transaction, error) {
	var q string = `SELECT tl.destination, tl.transaction_time, tl.transaction_id, COALESCE(tl.source_external,""), COALESCE(tl.source_internal,0), tl.amount, COALESCE(acc.name,"")
									FROM transaction_log tl
									LEFT JOIN account acc ON acc.account_id = tl.source_internal
									WHERE destination = ?
									ORDER BY transaction_time DESC`
	rows, err := db.Query(q, trans.DepositDest)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.DepositDest, &t.TransactionTime, &t.TransactionID, &t.ExternalSource, &t.InternalSource, &t.Amount, &t.Name)

		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (trans *Transaction) createTransaction(db *sql.DB) error {
	var q string

	var res sql.Result

	if trans.InternalSource == 0 {
		q = `INSERT INTO transaction_log
				(source_external, destination, amount)
				VALUES
				(?,?,?)`
		r, err := db.Exec(q, trans.ExternalSource, trans.DepositDest, trans.Amount)
		if err != nil {
			return err
		}
		res = r
	} else {
		q = `INSERT INTO transaction_log
				(source_internal, destination, amount)
				VALUES
				(?,?,?)`
		r, err := db.Exec(q, trans.InternalSource, trans.DepositDest, trans.Amount)
		if err != nil {
			return err
		}

		q = `SELECT email
				 FROM account
				 WHERE account_id = ?`
		db.QueryRow(q, trans.InternalSource).Scan(&trans.InternalSourceEmail)

		res = r
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	trans.TransactionID = uint64(id)
	return nil
}

func (trans *Transaction) getTransactions(db *sql.DB) ([]Transaction, error) {
	var q string = `SELECT tl.destination, tl.transaction_time, tl.transaction_id, COALESCE(tl.source_external,""), COALESCE(tl.source_internal,0), tl.amount, COALESCE(acc.name,"")
									FROM transaction_log tl
									LEFT JOIN account acc ON acc.account_id = tl.source_internal
									ORDER BY transaction_time DESC`
	rows, err := db.Query(q)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.DepositDest, &t.TransactionTime, &t.TransactionID, &t.ExternalSource, &t.InternalSource, &t.Amount, &t.Name)

		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
