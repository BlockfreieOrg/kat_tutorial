package main

import (
	"testing"
	"fmt"
	"database/sql"
	"io/ioutil"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

func TestPostiveBalance(t *testing.T) {
	entry := Entry{1, 1, 1}
	if !PositiveTransfer(entry)(nil) {
		t.Errorf("Expected postive balance")
	}
}

func TestSenderExists(t *testing.T) {
	dbOperation(func(tx *sql.Tx){
		entry := Entry{1, 1, 1}

		if !And(CreateSchema)(tx) {
			t.Errorf("could not create schema")
		}
		if SenderExists(entry)(tx){
			t.Errorf("did not expect sender")
		}

	})
}

func dbOperation(op func(*sql.Tx)) {
	tmpfile, err := ioutil.TempFile("", "kat_test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", tmpfile.Name())
	db, err := sql.Open("sqlite3",tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	op(tx)
	defer db.Close()

}

func TestCreateSchema(t *testing.T) {
	dbOperation(func(tx *sql.Tx){
		if !CreateSchema(tx) {
			t.Errorf("could not create schema")
		}
	})
}

/*
func PositiveTransfer(entry Entry)  KatExpression {
var CreateLedger = ExecuteSQL(`CREATE TABLE IF NOT EXISTS ledger
var DropLedger = ExecuteSQL("DROP TABLE IF EXISTS ledger")
var CreateQuarantine = ExecuteSQL(`CREATE TABLE IF NOT EXISTS quarantine
var DropQuarantine = ExecuteSQL("DROP TABLE IF EXISTS quarantine")
var CreateSchema = And(DropLedger,
func DumpLedger(tx *sql.Tx) bool {
func DumpQuarantine(tx *sql.Tx) bool {
var DumpState = And(DumpLedger,DumpQuarantine)
func UserExists(id int) KatExpression {
func SenderExists(entry Entry) KatExpression {
func RecieverExists(entry Entry) KatExpression {
func VerifyTransaction(entry Entry) KatExpression {
func UpdateLedger(id int,delta int) KatExpression {
func SaveTransaction(entry Entry)  KatExpression {
func CreateUser(id int) KatExpression {
func CreateSender(entry Entry) KatExpression {
func CreateReciever(entry Entry) KatExpression {
func QuarantineTransaction(entry Entry) KatExpression {
func UserBalancePositive(id int) KatExpression {
func SenderPositiveBalance(entry Entry) KatExpression {
func ReceiverPositiveBalance(entry Entry) KatExpression {
func EnsureSender(entry Entry) KatExpression {
func EnsureReciever(entry Entry) KatExpression {
func ProcessTransaction(entry Entry) KatExpression {
func ProcessBatch(entries []Entry) KatExpression {
func ProcessFile(path string) []Entry {
func main
*/
