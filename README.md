    First off, the description is a bit hard to digest.  You might
    want to expand on what you're doing more along with why it's
    interesting or important.  This has something to do with
    blockchain ledger in Kleene algebras?  You might also want to give
    a quick run down of the higher level concepts and why it's
    interesting to do them together?

    Under the "Kleene algebra" section, you might want to try to use a
    different symbol for multiplication so you can use the '*' symbol
    for the Kleene-star.  Maybe '.'?  Having the weird symbol is a bit
    confusing and hard to read.  Also consider using LaTex.  I don't
    think you can use it in GitHub Markdown on their site, but if you
    self host or host it on a GitHub pages page, then you could put in
    MathJax or something similar.

    When saying that the operations should be familiar if you have
    regular expression knowledge, you might want to expand (I know
    regex and I still don't quite see the parallel).  Maybe put it in
    a table with "Kleen operator" on the left and "Rexexp equivalent"
    on the right.

    Under "Motivation", you should mention that the code, structure,
    etc. is in Golang (maybe even provide a link).

    The picture is good but a little confusing.  I don't quite know
    what 'transactions' are, what the 'save batch', 'batch' etc and
    all the rest are.  It's kind of given without explanation.  What
    is quarantining?  What is 'dump state'?

    Under "0,K - operations ...", the sentence "A possible test if the
    check if a transaction has positive amount" is malformed..."A
    possible test is to check if the transaction amount is positive."?

    You might want to go into more detail about what you're doing
    here.  You're actually returning a function that tests the amount,
    not testing the amount.  Why?

    Under '* - sequential composition", but 'And' in bold or italics
    or with the `` quotes and do that wherever you refer to it so
    people know you're talking about the "and" operator.

    Under "+ - Choice" the sentence should be "Let's consider the
    following algebraic expressions".

    You should briefly mention what the salt is doing there.

    "This operation tries an operation.  If it fails, it rolls back the transaction and tries the next one."  and I don't quite know what you were trying to say with the last sentence.."It fails if all the operators return false the identify of the or operator"(?)

    Under "Fun with Algebra", 0*p = p?

  To build the go example, you need the 'github.com/rs/xid' package, so you might want to add that under the "Motivation" section.

  You might also want to tell how to use the go code.  I had trouble getting it to run when I cloned the repo:


# Kleene algebra of tests and transactions

Have you are given a simple business requirement.  Once you start
coding and refining these requirements, you quickly find yourself
overwhelmed tracking cases and conditions to guarantee and guard
against.  To handle these cases the code quickly devolves in to a
large nested if else then statements.  These complexity of statements
make them difficult to test and refactor safely.  Sounds familar?

This tutorial introduces an approach structuring code to deal with the
complexity of refining requirements and ensuring that the
implementation meets these requirements.

Lets consider the following requirement.

Check if the sender and receiver can enact the transaction, if so
do the transaction otherwise otherwise quarantine the transaction.


This article explores structuring of database
operations as Kleene algebra expressions.  It
demonstrates how to build an application out of
small units of code called tests and joint them
together using operators.

The code for this tutorial is available on GitHub.

[https://github.com/BlockfreieOrg/kat_tutorial](https://github.com/BlockfreieOrg/kat_tutorial)

Send me a link to your implementation in your favorite
language, and I will include a link.

# Kleene algebra

A Kleene algebra is an algebraic structure comprised
of operators and symbols.

```math
(K,+ , * ,  ⃰,0,1)
```

Symbols
   * K - K is a collection of operations that either succeed or fail.

	   Within this collection, there are a few _distinguished elements_ .

       * 0 - A trivial operation that fails.
       * 1 - A trivial operation that is successful.

Operators
   * a * b - sequential composition
   * a + b : choice
   * a ⃰  :  iteration


If you use regular expressions, these operators should
be familiar.

# Motivation - ledger

Let's consider the problem of maintaining a ledger.  This
application is a database-backed ledger processing batches
of transactions.

Transition entries specified by the following data structure

```go
type Entry struct {
     FromId         int `json:"FromId"`
     ToId           int `json:"ToId"`
     TransferAmount int `json:"TransferAmount"`
}
```

Transactions are provided in a file one transaction per line
in JSON format.  Below is a sample transaction entry.

```javascript
{"FromId":1,"ToId":2,"TransferAmount":10}
```

These entries are loaded from the file and processed by the
application.  If you have "Go" installed, you can build the
project.


```shell
go build
```

You can then run a sample transactions.  The operational argument
verbose allows you to see the SQL statements that are executed.

```shell
./kat_tutorial -dbfile=/tmp/tmp.db -infile=sample.json -verbose
```
It should apply the transactions to the ledger
and output the final state of ledger.

![load ledger](load-ledger.png)

The application first loads the transactions into a batch table. The
method batch entry then applies the transaction either updating the
ledger or quarantining the entry. The application then outputs the
state of the database.

# 0,K - operations that either succeed or fail

We can express a test as a function that takes a transaction and
returns a boolean. It is important to note, the tests themselves must
be stateless. The statefulness comes from the database, and the
database only.  The use of global variables, closures, or external
states is not permitted. For this implementation this is given by the
following type expression:

```go
type KatExpression func(*sql.Tx) bool
```

Let consider the symbol 0 (zero).  The implementation of zero
simply returns false which indicating the command has failed.

```go
func Zero(*sql.Tx) bool { return false }
```

A possible test if the check if a transaction has a postive amount.

```go
func PositiveTransfer(entry Entry)  KatExpression {
        return func(tx *sql.Tx) bool {
                return entry.TransferAmount > 0
        }
}
```

# 1,! - negation

So lets make a new expression out of old expression by introducing
the not operator.

```haskell
func Not(op KatExpression) (KatExpression) {
        return func(tx *sql.Tx) bool {
                return ! op(tx)
        }
}
```

Intuitively we know that negating 0 (zero) should give us 1 (one).

```math
1 = ¬ 0
```

The not operator which negates an expression and
can be used to create one from zero.

```go
var One KatExpression = Not(Zero)
```

# * - sequential composition

Typically when a transaction is received, it is verified. First, it
checks the transfer amount is positive and then verify that the sender
and the receiver exist.

```math
VerifyTransaction = PositiveTransfer * SenderExists * RecieverExists
```

This introduces the need for the * operator we shall refer to as And.

```go
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
```

The and operator behaves as we would expect an and operator to
work. It first tries the first operator, If this returns true, it then
tries the next term. If any of the operators return false, it returns
false otherwise it returns true.

We can implement the pseudo code from the beginning of the section.

```go
func VerifyTransaction(entry Entry) go_kat.KatExpression {
        return go_kat.And(PositiveTransfer(entry),
		          SenderExists(entry),
		          RecieverExists(entry))
}
```

# + - Choice

Lets consider the follow algebraic expressions

```math
ensureSender = !senderExists + createSender

ensureReciever = !recieveExists + createReciever
```

If the sender/reciever does not exists then create the sender/reciever.

To implement the above expression, let's provide the Or operator.

```go
func Or(args ... KatExpression) (KatExpression) {
	return func(tx *sql.Tx) bool {
		result := true
		var salt = xid.New()
		tx.Exec(fmt.Sprintf("savepoint %s", salt))
		for _, op := range args {
			result = op(tx)
			if result {
				break
			} else {
				tx.Exec(fmt.Sprintf("rollback to savepoint %s", salt))
			}
		}
		return result
	}
}
```

This operator tries an operation if it fails it rolls back
the transaction and tries the next one.  It fails if all
the operators return false the identity of the or operator.


Below is the implementation of the expression "ensureSender".

```go
func EnsureSender(entry Entry) KatExpression {
	return Or(SenderExists(entry),
		         CreateSender(entry))
}
```

#      ⃰- Star

Proccessing a transaction is described by the following equation.

```math
processEntry = removeBatch
               * ((ensureSender
                  * ensureReciever
                  * verifyTransaction
                  * saveTransaction
                  * senderPositiveBalance
                  * receiverPositiveBalance)
                 + quarantineTransaction)
```

The implementation and details of the functions in the expression are
omitted for brevity.  The expression first checks the transaction
applies the transaction then checks the result. If a transaction
fails, it is placed in quarantine. We then want to process each
transaction for each entry.


```math
batchEntry = processEntry ⃰
```

We then need a star operator.  This operator applies an expression
until it returns false.

```go
func Star(op KatExpression) (KatExpression) {
	return func(tx *sql.Tx) bool {
		for {
			var salt = xid.New()
			tx.Exec(fmt.Sprintf("savepoint %s", salt))
			if (!op(tx)){
				tx.Exec(fmt.Sprintf("rollback %s", salt))
				return true
			}
		}

	}
}
```
There is little helper function that is needed to load
a batch entry and then process the transaction.

```go
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
```

The expression procces entry is implemented as follows

```go
func ProcessEntry(id int,entry Entry) KatExpression {
	return And(And(BatchExists(id),DeleteBatch(id),Not(BatchExists(id))),
	           Or(And(EnsureSender(entry),
	                  EnsureReciever(entry),
	                  VerifyTransaction(entry),
	                  SaveTransaction(entry),
	                  SenderPositiveBalance(entry),
			  ReceiverPositiveBalance(entry)),
	              QuarantineTransaction(entry)))
}
```

The entries are processed by the simple expression

```go
var BatchEntry = Star(ProcessBatch(ProcessEntry))
```

# Fun with Algebra

Using this approach one ends up with small blocks of code held
together with operators.  Note the code is absent the error checking
code and conditional code which tend to increase complexity.
Formulating an operation as a Kleene algebraic expression
allows for algebraically restructuring. That's to say also long as
two expressions are algebraically equivalent they are (and do) the
same things.  Here are a few identities to get started.

```math
0 + p = p
p + p = p
1 * p = p
p(q + r) = p*q + p*r
0 * p = p
1 + p p ⃰ = p ⃰
```

Consider the following simplifications:

```math
Zero = Zero * ensureSender

ensureSender = One * ensureSender
```

These equalities follow from algebra.

```math
ensureSender * ensureReciever =
((!senderExists + createSender) * !recieveExists)
+
((!senderExists + createSender) * createReciever)
```

So with very little examination we know that these two expressions are
equivalent.  So this gives a way to quick restructure code without much
effort.

# Hoare triples

A pattern to apply to expressions is the Hoare triple.

```math
   {P} * C *  {Q}
```

   * P - is a precondition
   * C - command
   * Q - is a postcondition

The precondition P guards the command C. The postcondition Q is
checked after executing C to ensure the desired side effects.

An example of this is _removeBatch_.

```math
removeBatch = batchExists * deleteBatch * !batchExists
```

This first test checks if the batch entry exists. Then the entry is
removed by command _deleteBatch_ removing it from the database. It
then checks for the removal of the batch entry. These pre and
postconditions ensure that remove batch behaves as expected.

Another example is a better _createSender_
that has the requisite pre and postconditional tests for creating
a sender.

```math
betterCreateSender = !senderExists * createSender * senderExists
```
# Further Reading and Sources

## [Kleene Algebra with Tests: A Tutorial](https://www.cl.cam.ac.uk/events/ramics13/KozenTutorial1.pdf)

This presentation is a slight variation on the standard treatment of
Kleene Algebra with tests. The difference being is it difficult in an
imperative language to make tests with guarantees of no side effects
other than that of the database. As a result, the tests and the
commands they unified in this approach.


## [A Short Introduction to Hoare Logic](https://www.cse.iitb.ac.in/~supratik/courses/cs615/msri_ss08.pdf)

Structuring the tests is key to building these expressions. This
discussion on Hoare logic gives an excellent overview of how to
structure them.
