# Cash Deposit Backend REST API

### A Cash Deposit Backend REST API based on Go and MySQL included with Unit Test
----------
## Introduction
Features:
1. New account registration
2. Cash deposit from customer and non-customer to the existing account
3. View total balance and details of an existing account
4. View deposit history of an account
5. Email notification for every deposit made via the app

Dependencies:
1. [godotenv](https://github.com/joho/godotenv)
2. [mysql driver](https://github.com/go-sql-driver/mysql)
3. [gorilla/mux router](https://github.com/gorilla/mux)
4. [gorilla/handlers](https://github.com/gorilla/handlers)

Design Documents:
- [Google Drive](https://drive.google.com/drive/folders/1u6Mjjt-G1yT-WPKTVegYorq0hi2tmDP5?usp=sharing)
 
----------
## How to use
0. You should already have ```Go``` installed in your local machine/server.
1. Before using this app, first you need to install all the dependencies stated above using ```go get```
2. Rename .env.example to .env and configure the variables there to match your environment.

### Development
Run:
```
go run main.go app.go model-account.go model-transaction.go
```

### Production
Run:
```
go build
```
and then execute the generated binary file

### Unit test
What's being tested?
1. Emptying table
2. Getting a non existent account
3. Getting an account and its balance
4. Creating a duplicate account
5. Creating account with invalid ID card number
6. Creating an account with invalid name
7. Creating an account with invalid email address
8. Creating a deposit from a registered account (email test included)
9. Creating a deposit from an unregistered account (email test included)
10. Creating a deposit with an invalid non-account email address
11. Creating a deposit with invalid account number as its source
12. Creating a deposit with invalid account number as its destination
13. Creating a deposit with invalid amount

Run:
```
go test -v
```

----------
 
## API Usage
## Index
#### Account
1. [Get All Accounts](#get-all-accounts)
2. [Get an account details](#get-an-account-details)
3. [Register A New Account](#register-a-new-account)

#### Transaction
1. [Get all deposit history](#get-all-deposit-history)
2. [Get deposit history of an account](#get-deposit-history-of-an-account)
3. [Create a new deposit](#create-a-new-deposit)

## Account
### Get all accounts
Get a list of all accounts

**Method and path**
```
  GET /account
```
**Request Parameters**

| Property | Type  | Required | Description |
| -------- | ----- | -------- | ----------- |
| -        | -     | -        | -           |

**Example**

**Request**
```
  GET /account
```
**Response**

**Success**  
HTTP Response code: ```200 OK```
```json
[
  {
    "accountid": 3,
    "idcardno": "1231234567890123",
    "name": "Chris",
    "email": "chris@mail.com",
    "balance": 14000
  },
  {
    "accountid": 1,
    "idcardno": "1234567890123456",
    "name": "James",
    "email": "james@mail.com",
    "balance": 127000
  },
  {
    "accountid": 2,
    "idcardno": "1234567890987654",
    "name": "Bryan",
    "email": "bryan@mail.com",
    "balance": 1239000
  }
]
```
**Fail**  
HTTP Response code: ```200 OK```
```json
[]
```
----------
### Get an account details

Get the details of an existing account

**Method and path**
```
  GET /account/{account_id}
```
**Request Parameters**

| Property | Type  | Required | Description |
| -------- | ----- | -------- | ----------- |
| -        | -     | -        | -           |

**Example**

**Request**
```
  GET /account/3
```
**Response**

**Success**  
HTTP Response code: ```200 OK```
```json
  {
    "accountid": 3,
    "idcardno": "1231234567890123",
    "name": "Chris",
    "email": "chris@mail.com",
    "balance": 14000
  }
```
**Fail**  
HTTP Response code: ```404 Not Found```
```json
  {
    "error": "Account not found"
  }
```
----------
### Register a new account
Create a new account

**Method and path**
```
  POST /account
```
**Request Parameters**

| Property    | Type     | Required   | Description                                             |
| ----------- | -------- | ---------- | ------------------------------------------------------- |
| idcardno    | string   | Yes        | A valid Indonesian ID card number (16 characters long)  |
| name        | string   | Yes        | Name of the customer                                    |
| email       | string   | Yes        | A valid customer email address                          | 

**Example**

**Request**
```json
  POST /account

  {
    "idcardno":"1234567890123456",
    "name":"John Doe",
    "email":"john.doe@mail.com"
  }
```
**Response**

**Success**  
HTTP Response code: ```201 Created```
```json
  {
    "accountid": 10,
    "idcardno": "1234567890123456",
    "name": "John Doe",
    "email": "john.doe@mail.com",
    "balance": 0
  }
```
**Fail**  
HTTP Response code: ```400 Bad Request```
```json
  {
    "error": "Account exists!"
  }
```
**or**
```json
  {
    "error": "Invalid ID Card"
  }
```
**or**
```json
  {
    "error": "Invalid Name"
  }
```
**or**
```json
  {
    "error": "Invalid Email Address"
  }
```
----------
## Transaction
### Get all deposit history
Get a list of all deposit history

**Method and path**
```
  GET /transaction
```
**Request Parameters**

| Property | Type  | Required | Description |
| -------- | ----- | -------- | ----------- |
| -        | -     | -        | -           |

**Example**

**Request**
```json
  GET /transaction
```
**Response**

**Success**  
HTTP Response code: ```200 OK```
```json
[
  {
    "transid": 3,
    "depositdest": 2,
    "externalsource": "",
    "internalsource": 2,
    "internalsourceemail": "bryan@mail.com",
    "name": "Bryan",
    "amount": 1250000,
    "transtime": "2018-09-18T09:24:32Z"
  },
  {
    "transid": 2,
    "depositdest": 1,
    "externalsource": "levy@mail.com",
    "internalsource": 0,
    "internalsourceemail": "",
    "name": "",
    "amount": 1350000,
    "transtime": "2018-09-17T20:24:43Z"
  },
  {
    "transid": 1,
    "depositdest": 1,
    "externalsource": "",
    "internalsource": 1,
    "internalsourceemail": "james@mail.com",
    "name": "James",
    "amount": 1500000,
    "transtime": "2018-09-17T20:19:57Z"
  }
]
```
**Fail**  
HTTP Response code: ```200 OK```
```json
  []
```
----------

### Get deposit history of an account
Get a cash deposit history of an account

**Method and path**
```
  GET /account/{account_id}/history
```
**Request Parameters**

| Property | Type  | Required | Description |
| -------- | ----- | -------- | ----------- |
| -        | -     | -        | -           |

**Example**

**Request**
```json
  GET /account/3/history
```
**Response**

**Success**  
HTTP Response code: ```200 OK```
```json
[
  {
    "transid": 26,
    "depositdest": 3,
    "externalsource": "",
    "internalsource": 1,
    "internalsourceemail": "james@mail.com",
    "name": "James",
    "amount": 193000,
    "transtime": "2018-09-19T15:34:02Z"
  },
  {
    "transid": 25,
    "depositdest": 3,
    "externalsource": "michael@mail.com",
    "internalsource": 0,
    "internalsourceemail": "",
    "name": "",
    "amount": 1253000,
    "transtime": "2018-09-19T15:32:30Z"
  }
]
```
**Fail**  
HTTP Response code: ```200 OK```
```json
  []
```
----------

### Create a new deposit
Create a new cash deposit to an existing account

**Method and path**
```
  POST /transaction
```
**Request Parameters**

| Property       | Type        | Required   | Description                    |
| -------------- | ----------- | ---------- | ------------------------------ |
| depositdest    | int         | Yes        | A valid account id             |
| externalsource | string      | Yes*       | Email address of the depositor |
| internalsource | int         | Yes*       | A valid account id             | 
| amount         | int/float   | Yes        | Deposit amount (in Rupiah)     |
*) Only one of ```externalsource``` or ```internalsource``` should present in the request parameters

**Example**

**Request**
```json
  POST /transaction

  {
    "depositdest":2,
    "externalsource":"brown@mail.com",
    "amount":1325000.00
  }
```
**or**
```json
  POST /transaction

  {
    "depositdest":2,
    "internalsource":1,
    "amount":8970000
  }
```

**Response**

**Success**  
HTTP Response code: ```201 Created```
```json
  {
    "transid": 34,
    "depositdest": 2,
    "externalsource": "brown@mail.com",
    "internalsource": 0,
    "internalsourceemail": "",
    "name": "",
    "amount": 1325000,
    "transtime": "2018-09-20T20:19:19Z"
  }
```
**or**
```json
  {
    "transid": 35,
    "depositdest": 2,
    "externalsource": "",
    "internalsource": 1,
    "internalsourceemail": "james@mail.com",
    "name": "James",
    "amount": 8970000,
    "transtime": "2018-09-20T20:21:19Z"
  }
```

**Fail**  
HTTP Response code: ```400 Bad Request```
```json
  {
    "error": "Invalid request payload"
  }
```
**or**
```json
  {
    "error": "Invalid Account ID"
  }
```
**or**
```json
  {
    "error": "Invalid Email Address"
  }
```
----------
