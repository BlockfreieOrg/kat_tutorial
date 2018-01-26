package main

import ("os"
	"database/sql"
	"fmt"
	"flag"
	"log"
	"bufio"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	FromId         int `json:"FromId"`
	ToId           int `json:"ToId"`
	TransferAmount int `json:"TransferAmount"`
}

func PositiveTransfer(entry Entry)  KatExpression {
	return func(tx *sql.Tx) bool {
		return entry.TransferAmount > 0
	}
}
/* ledger */
var CreateLedger = ExecuteSQL(`CREATE TABLE IF NOT EXISTS ledger
                                         (UserId integer,
                                          UserBalance integer)`)
var DropLedger = ExecuteSQL("DROP TABLE IF EXISTS ledger")

/* batch */
var CreateBatch = ExecuteSQL(`CREATE TABLE IF NOT EXISTS batch
                                     (Id integer primary key autoincrement,
                                      FromId integer,
                                      ToId integer,
                                      TransferAmount integer)`)
var DropBatch = ExecuteSQL("DROP TABLE IF EXISTS batch")

/* quarantine */
var CreateQuarantine = ExecuteSQL(`CREATE TABLE IF NOT EXISTS quarantine
                                     (FromId integer,
                                      ToId integer,
                                      TransferAmount integer)`)
var DropQuarantine = ExecuteSQL("DROP TABLE IF EXISTS quarantine")


var CreateSchema = And(DropLedger,
	               CreateLedger,
	               DropBatch,
	               CreateBatch,
	               DropQuarantine,
 	               CreateQuarantine)

func DumpBatch(tx *sql.Tx) bool {
	fmt.Printf("Batch\n")
	var Id = -1
	var FromId = -1
	var ToId = -1
	var TransferAmount = -1
	var handler = func(){
		fmt.Printf("{ Id: %v , FromId: %v , ToId : %v , TransferAmount : %v }\n",
			  Id,
			  FromId,
	                  ToId,
			  TransferAmount)
	}
	var sql = `select Id,
			  FromId,
	                  ToId,
			  TransferAmount
                    from batch`
	return HandleQuery(sql)(tx,handler,&Id,&FromId,&ToId,&TransferAmount)
}


func DumpLedger(tx *sql.Tx) bool {
	fmt.Printf("Ledger\n")
	var UserId = -1
	var UserBalance = -1
	var handler = func(){
		fmt.Printf("{ UserId: %v , UserBalance %v }\n",
			   UserId,
			   UserBalance)
	}
	var sql = `select UserId ,
                   UserBalance from ledger`
	return HandleQuery(sql)(tx,handler,&UserId,&UserBalance)
}

func DumpQuarantine(tx *sql.Tx) bool {
	fmt.Printf("Quarantine\n")
	var from_id = -1
	var to_balance = -1
	var transfer_amount = -1
	var handler = func(){
		fmt.Printf("{ from_id : %v , to_balance : %v , transfer_amount : %v }\n",
                           from_id,
			   to_balance,
			   transfer_amount)
	}
	var sql = `select FromId,
			   ToId,
			   TransferAmount
                   from quarantine`
	return HandleQuery(sql)(tx,handler,&from_id,&to_balance,&transfer_amount)
	return true
}

var DumpState = And(DumpBatch,DumpLedger,DumpQuarantine)

func UserExists(id int) KatExpression {
	return func(tx *sql.Tx) bool {
		var result = false
		var sql = "SELECT count(*) > 0 FROM ledger WHERE UserId=?"
		result = ExecuteQuery(sql,id)(&result)(tx) && result
		return  result
	}
}


func SenderExists(entry Entry) KatExpression {
	return UserExists(entry.FromId)
}

func RecieverExists(entry Entry) KatExpression {
	return UserExists(entry.ToId)
}


func VerifyTransaction(entry Entry) KatExpression {
        return And(PositiveTransfer(entry),
		          SenderExists(entry),
		          RecieverExists(entry))
}


func UpdateLedger(id int,delta int) KatExpression {
	var sql = `UPDATE ledger
                   SET UserBalance = UserBalance + ?
                   WHERE UserId =?`
	return ExecuteSQL(sql,delta,id)
}

func SaveTransaction(entry Entry)  KatExpression {
	return And(UpdateLedger(entry.FromId,-entry.TransferAmount),
		          UpdateLedger(entry.ToId,entry.TransferAmount))
}

func CreateUser(id int) KatExpression {
	var sql = `INSERT INTO ledger
                   (UserId,UserBalance)
                   VALUES
                   (?,100)`
	return ExecuteSQL(sql,id)
}



func CreateSender(entry Entry) KatExpression {
	return CreateUser(entry.FromId)
}

func CreateReciever(entry Entry) KatExpression {
	return CreateUser(entry.ToId)
}

func QuarantineTransaction(entry Entry) KatExpression {
	var sql = `INSERT INTO quarantine
                   (FromId,ToId,TransferAmount)
                   VALUES
                   (?,?,?)`
	return ExecuteSQL(sql,
		                 entry.FromId,
		                 entry.ToId,
		                 entry.TransferAmount)
}

func UserBalancePositive(id int) KatExpression {
	return func(tx *sql.Tx) bool {
		var result = false
		var sql = "SELECT UserBalance > 0 FROM ledger WHERE UserId=?"
		return ExecuteQuery(sql,id)(&result)(tx) && result
	}
}

func SenderPositiveBalance(entry Entry) KatExpression {
	return UserBalancePositive(entry.FromId)
}

func ReceiverPositiveBalance(entry Entry) KatExpression {
	return UserBalancePositive(entry.ToId)
}

func EnsureSender(entry Entry) KatExpression {
	return Or(SenderExists(entry),
		         CreateSender(entry))
}

func EnsureReciever(entry Entry) KatExpression {
	return Or(RecieverExists(entry),
		         CreateReciever(entry))
}

func SaveBatch(entries []Entry) KatExpression {
	return func(tx *sql.Tx) bool {
		for _, entry := range entries {
			var sql = `INSERT INTO batch
                                   (FromId,ToId,TransferAmount)
                                   VALUES
                                   (?,?,?)`
			if ! ExecuteSQL(sql,entry.FromId,entry.ToId,entry.TransferAmount)(tx) {
				return false
			}
		}
		return true
	}
}

func DeleteBatch(id int) KatExpression {
	var sql = `DELETE FROM batch WHERE id = ?`
	return ExecuteSQL(sql,id)
}

func BatchExists(id int) KatExpression {
	return func(tx *sql.Tx) bool {
		var result = false
		var sql = "SELECT count(*) > 0 FROM batch WHERE id=?"
		result = ExecuteQuery(sql,id)(&result)(tx) && result
		return  result
	}
}

func RemoveBatch(id int) KatExpression {
	return And(BatchExists(id),DeleteBatch(id),Not(BatchExists(id)))
}

func ProcessEntry(id int,entry Entry) KatExpression {
	return And(RemoveBatch(id),
	           Or(And(EnsureSender(entry),
	                  EnsureReciever(entry),
	                  VerifyTransaction(entry),
	                  SaveTransaction(entry),
	                  SenderPositiveBalance(entry),
			  ReceiverPositiveBalance(entry)),
	              QuarantineTransaction(entry)))
}

func ProcessBatch(op (func(int,Entry) KatExpression)) KatExpression {
	return func(tx *sql.Tx) bool {
		var id , fromId , toId, transferAmount = -1, -1, -1, -1
		var sql = `SELECT Id, FromId, ToId, TransferAmount
                           FROM batch`
		var result = ExecuteQuery(sql,id)(&id,&fromId,&toId,&transferAmount)(tx)
		if result {
			result = op(id,Entry{fromId , toId, transferAmount})(tx)
		}
		return result
	}
}

var BatchEntry = Star(ProcessBatch(ProcessEntry))

func ProcessFile(path string) []Entry {

	var result []Entry
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Entry
		text := []byte(scanner.Text())
		err := json.Unmarshal(text, &e)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("%v\n", e)
			result = append(result, e)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func main() {
	var inFileFlagPtr = flag.String("infile", "", "in file")
	var dbFileFlagPtr = flag.String("dbfile", "", "db file")
	var dbVerbosePtr = flag.Bool("verbose", false, "verbose ")

	flag.Parse()
	fmt.Println("infile:", *inFileFlagPtr)
	fmt.Println("dbfile:", *dbFileFlagPtr)
	fmt.Println("verbose:", *dbVerbosePtr)

	if(!*dbVerbosePtr){
		LogMessage = func(msg string){}
	}

	var ops = And(CreateSchema,
		      SaveBatch(ProcessFile(*inFileFlagPtr)),
		      BatchEntry,
		      DumpState)
	Eval("sqlite3",*dbFileFlagPtr,ops)
}
