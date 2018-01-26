package main

import (
	"fmt"
	"database/sql"
	"log"
	"github.com/rs/xid"
)

type KatExpression func(*sql.Tx) bool

var ( LogMessage (func(string)) = func(msg string){
		log.Print(msg)
	}

	LogError (func(error) bool) = func(err error) bool {
		if err != nil {
			log.Print(err)
			return false
		} else {
			return true
		}
	}
)

func Zero(*sql.Tx) bool { return false }

func And(args ... KatExpression) (KatExpression) {
	return func(tx *sql.Tx) bool {
		for _, op := range args {
			if (!op(tx)){
				return false
			}
		}
		return true
	}
}

func Or(args ... KatExpression) (KatExpression) {
	return func(tx *sql.Tx) bool {
		result := true
		var salt = xid.New()
		LogMessage(fmt.Sprintf("savepoint %s", salt))
		tx.Exec(fmt.Sprintf("savepoint %s", salt))
		for _, op := range args {
			result = op(tx)
			if result {
				break
			} else {
				LogMessage(fmt.Sprintf("rollback to savepoint %s", salt))
				tx.Exec(fmt.Sprintf("rollback to savepoint %s", salt))
			}
		}
		return result
	}
}

func Star(op KatExpression) (KatExpression) {
	return func(tx *sql.Tx) bool {
		for {   var salt = xid.New()
			LogMessage(fmt.Sprintf("savepoint %s", salt))
			tx.Exec(fmt.Sprintf("savepoint %s", salt))
			if (!op(tx)){
				LogMessage(fmt.Sprintf("rollback to savepoint %s", salt))
				tx.Exec(fmt.Sprintf("rollback %s", salt))
				return true
			}
		}

	}
}

func Not(op KatExpression) (KatExpression) {
	return func(tx *sql.Tx) bool {
		return ! op(tx)
	}
}

var One KatExpression = Not(Zero)

func Eval(driverName string, dataSourceName string,expression KatExpression) {
	db, err := sql.Open(driverName, dataSourceName)
	LogError(err)
	tx, err := db.Begin()
	defer db.Close()
	if LogError(err) {
		if expression(tx) {
			LogError(tx.Commit())
		} else {
			LogError(tx.Rollback())
		}
	}
}

func ExecuteSQL(statement string, args ...interface{}) KatExpression {
	return func(tx *sql.Tx) bool {
		LogMessage(fmt.Sprintf("ExecuteSQL: %v %v\n", statement,args))
		_, err := tx.Exec(statement,args...)
		return LogError(err)
	}
}

func ExecuteQuery(query string, args ...interface{}) func(...interface{}) KatExpression {
	return func(dest ...interface{}) KatExpression {
		return func(tx *sql.Tx) bool {
			rows, err := tx.Query(query,args...)
			LogMessage(fmt.Sprintf("ExecuteQuery: %v %v\n", query,args))
			if err != nil {
				return LogError(err)
			}
			defer rows.Close()
			if rows.Next() {
				if err := rows.Scan(dest...); err != nil {
					return LogError(err)
				}
			} else {
				return LogError(rows.Err()) && false
			}
			return LogError(rows.Err())
		}
	}
}

func HandleQuery(query string, args ...interface{}) func(*sql.Tx,func(),...interface{}) bool {
	return func(tx *sql.Tx,handler func(),dest ...interface{}) bool {
		LogMessage(fmt.Sprintf("HandleQuery: %v %v\n", query,args))

		rows, err := tx.Query(query,args...)
		if err != nil {
			return LogError(err)
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(dest...); err != nil {
				return LogError(err)
			} else {
				handler()
			}
		}
		return LogError(rows.Err())
	}
}
