package main

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"os"
	_ "fmt"
	_ "log"
	"io/ioutil"
	"database/sql"
)

func TestZero(t *testing.T) {
	if Zero(nil) {
		t.Errorf("zero not false")
	}
}

func TestOne(t *testing.T) {
	if !One(nil) {
		t.Errorf("one not true")
	}
}

func TestNot(t *testing.T) {
	if One(nil) != Not(Zero)(nil) {
		t.Errorf("one not not zero")
	}
}

func WithTestFile(t *testing.T,test func(string)){
	tmpfile, err := ioutil.TempFile("", "kat_test")
	if err != nil {
		t.Errorf("could not create temp file")
	}
	test(tmpfile.Name())
	if _, err := os.Stat(tmpfile.Name()); os.IsNotExist(err) {
		t.Error("could not find temp file")
	}
}

func WithTestExpression(t *testing.T,test KatExpression){
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,test)
	})
}

func assertExpression(t *testing.T,msg string,k KatExpression) KatExpression {
	return func(tx *sql.Tx) bool {
		result := k(tx)
		if !result {
			t.Error(msg)
		}
		return result
	}
}

func TestCanEvalExpression(t *testing.T) {
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",
		     tmpfile,
		     assertExpression(t,"evaluate expression : expected true",One))
	})
}

func TestCanExecuteSQL(t *testing.T) {
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"execute sql : drop a",Not(ExecuteSQL("drop table a"))))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute sql : create a",ExecuteSQL("create table a(a integer)")))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute sql : drop a",ExecuteSQL("drop table a")))
	})
}

func TestCanExecuteQuery(t *testing.T) {
	var hasOne = func(tx *sql.Tx) bool {
		var result = false
		var sql = "SELECT count(*) > 0 FROM a"
		return ExecuteQuery(sql)(&result)(tx) && result
	}

	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"execute query : 1 drop a",Not(ExecuteSQL("drop table a"))))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute query : create a",ExecuteSQL("create table a(a integer)")))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute query : 1 check 1",Not(hasOne)))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute query : insert 1",ExecuteSQL("insert into a (a) VALUES (1)")))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute query : 2 check 1",hasOne))
		Eval("sqlite3",tmpfile,assertExpression(t,"execute query : 2 drop a",ExecuteSQL("drop table a")))
	})
}

func TestCanHandleQuery(t *testing.T) {
	var sumThree = func(tx *sql.Tx) bool {
		var Tmp = 0
		var Count = 0
		var handler = func(){
			Count += Tmp
		}
		var sql = "select a from a"
		return HandleQuery(sql)(tx,handler,&Tmp) && Count == 3
	}

	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : 1 drop a",Not(ExecuteSQL("drop table a"))))
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : create a",ExecuteSQL("create table a(a integer)")))
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : 1 check sum 3",Not(sumThree)))
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : insert 1",ExecuteSQL("insert into a (a) VALUES (1)")))
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : insert 2",ExecuteSQL("insert into a (a) VALUES (2)")))
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : 2 check sum 3",sumThree))
		Eval("sqlite3",tmpfile,assertExpression(t,"handle query : 2 drop a",ExecuteSQL("drop table a")))
	})

}

func TestAndLogic(t *testing.T) {
	assertExpression(t,"logic : and",And())(nil)
	assertExpression(t,"logic : and 1",And(One))(nil)
	assertExpression(t,"logic : and 0",Not(And(Zero)))(nil)
	assertExpression(t,"logic : and 1 1",And(One,One))(nil)
	assertExpression(t,"logic : and 0 1",Not(And(Zero,One)))(nil)
	assertExpression(t,"logic : and 1 0",Not(And(One,Zero)))(nil)
	assertExpression(t,"logic : and 0 0",Not(And(Zero,Zero)))(nil)
}

func TestAndTransaction(t *testing.T) {
	var insertOne = assertExpression(t,"and insert 1",ExecuteSQL("insert into b (b) VALUES (1)"))
	var insertTwo = assertExpression(t,"and insert 2",ExecuteSQL("insert into b (b) VALUES (2)"))
	var checkSum = func(expected int) KatExpression {
		return func(tx *sql.Tx) bool {
			var Tmp = 0
			var Count = 0
			var handler = func(){
				Count += Tmp
			}
			var sql = "select b from b"
			return  HandleQuery(sql)(tx,handler,&Tmp) && Count == expected
		}
	}
	var createTable = And(ExecuteSQL("drop table if exists b"),ExecuteSQL("create table b (b integer)"))
	var dropTable = ExecuteSQL("drop table b")

	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"expected true",And(createTable,checkSum(0),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"expected true",And(createTable,insertOne,checkSum(1),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"expected true",And(createTable,insertOne,insertTwo,checkSum(3),dropTable)))
	})
}

func TestOrLogic(t *testing.T) {
	WithTestExpression(t,assertExpression(t,"logic : or",Or()))
	WithTestExpression(t,assertExpression(t,"logic : or 0",Or(One)))
	WithTestExpression(t,assertExpression(t,"logic : or 1",Not(Or(Zero))))
	WithTestExpression(t,assertExpression(t,"logic : or 1 1",Or(One,One)))
	WithTestExpression(t,assertExpression(t,"logic : or 1 0",Or(Zero,One)))
	WithTestExpression(t,assertExpression(t,"logic : or 1 0",Or(One,Zero)))
	WithTestExpression(t,assertExpression(t,"logic : or 0 0",Not(Or(Zero,Zero))))
}

func TestOrTransaction(t *testing.T) {
	var insertOne = assertExpression(t,"or failed insert 1",ExecuteSQL("insert into b (b) VALUES (1)"))
	var insertTwo = assertExpression(t,"or failed insert 2",ExecuteSQL("insert into b (b) VALUES (2)"))

	var checkSum = func(expected int) KatExpression {
		return func(tx *sql.Tx) bool {
			var Tmp = 0
			var Count = 0
			var handler = func(){
				Count += Tmp
			}
			var sql = "select b from b"
			return  HandleQuery(sql)(tx,handler,&Tmp) && Count == expected
		}
	}
	var createTable = And(ExecuteSQL("drop table if exists b"),ExecuteSQL("create table b (b integer)"))
	var dropTable = ExecuteSQL("drop table b")
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or noop",And(createTable,checkSum(0),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or empty",And(createTable,Or(),checkSum(0),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or zero",Not(And(createTable,Or(Zero),checkSum(0),dropTable))))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or +One",And(createTable,Or(insertOne),checkSum(1),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or O + One",And(createTable,Or(Zero,insertOne),checkSum(1),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or +One 0",And(createTable,Or(insertOne,Zero),checkSum(1),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or +One +Two",And(createTable,Or(insertOne,insertTwo),checkSum(1),dropTable)))
	})
	WithTestFile(t,func(tmpfile string){
		Eval("sqlite3",tmpfile,assertExpression(t,"t or +Two +One",And(createTable,Or(insertTwo,insertOne),checkSum(2),dropTable)))
	})
}

func TestStarTransaction(t *testing.T) {
	var insertOne = assertExpression(t,"or failed insert 1",ExecuteSQL("insert into s (s) VALUES (1)"))
	var checkSum = func(expected int) KatExpression {
		return func(tx *sql.Tx) bool {
			var Tmp = 0
			var Count = 0
			var handler = func(){
				Count += Tmp
			}
			var sql = "select s from s"
			return  HandleQuery(sql)(tx,handler,&Tmp) && Count == expected
		}
	}
	var insertn = func(excepted int) KatExpression {
		var Count = 0
		return func(tx *sql.Tx) bool {
			Count = Count + 1
			if Count <= excepted {
				return insertOne(tx)
			} else {
				return Zero(tx)
			}
		}
	}
	WithTestExpression(t,assertExpression(t,"start : empty",Star(Zero)))
	var createTable = And(ExecuteSQL("drop table if exists b"),ExecuteSQL("create table s (s integer)"))
	var dropTable = ExecuteSQL("drop table s")
	WithTestExpression(t,assertExpression(t,"star : empty",And(createTable,checkSum(0),dropTable)))
	WithTestExpression(t,assertExpression(t,"star : 5",And(createTable,checkSum(0),Star(insertn(5)),checkSum(5),dropTable)))
}
