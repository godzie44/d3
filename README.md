![golangci-lint](https://github.com/godzie44/d3orm/workflows/golangci-lint/badge.svg) ![tests](https://github.com/godzie44/d3orm/workflows/tests/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/godzie44/d3/badge.svg?branch=master)](https://coveralls.io/github/godzie44/d3?branch=master)

# D3 ORM

## GOLANG ORM for DDD applications

D3 is golang ORM and DataMapper. This project was design with respect to such 
ORM's like hybernate and doctrine. Main task - give an instrument for create 
nice domain layer. If your business code not expressive enough, or if you want unit 
tests without a database, or if in your application, for some reason, you need aggregates, 
repositories, entities and value objects - d3 can be a good choice.

## Motivation. Why another ORM?

In my appinion in GO have a lot of good ORM's. They are pretty fast and 
save the developer from a lot of boilerplate code. But, if you want write code
in DDD style (using DDD patterns and philosofy) ORM's don't help, 
because DDD approach imposes certain requirements. Main requirement - 
persistence ignorance. And current GO ORM's do not provide sufficient level of abstraction for this.
Other words, we need keep business logic free of data access code. That's why D3 created.

## Main futures

- code generation instead of reflection
- one-to-one, one-to-many and many-to-many relations
- lazy and eager loading
- fetch strategies (eager/lazy as above or extract with relations in one joined query)
- query builder
- db schema auto generation
- first level cache
- cascade remove of related entities
- UUID support
- application level transaction (UnitOfWork)
- transactions support
- smart persist layer dont generate redundant queries on entity updates

## Documentation

### Quick start

Here is simple example of usage:

0. Install d3
```go
    go get -u github.com/godzie44/d3
```
1. Create entity, and generate code with d3 command (d3 <file>.go or d3 <directory>).
```go
    //d3:entity
    //d3_table:user
    type user struct {
    	id     sql.NullInt32      `d3:"pk:auto"`
    	name   string 
    }
```
2. Create db connection, orm instance, register entities.
```go
	pgDb, _ := pgx.Connect(context.Background(), os.Getenv("DB"))
	d3orm := orm.NewOrm(d3pgx.NewPgxDriver(pgDb, &adapter.SquirrelAdapter{}))
	_ = d3orm.Register(&user{})
```
3. Create session and repository.
```go
    ctx := d3orm.CtxWithSession(context.Background())
    rep, _ := d3orm.MakeRepository(&user{})
```
4. Find existing user by id.
```go
    existingUser, err := rep.FindOne(ctx, rep.MakeQuery().AndWhere("id = ?", 1))
```
5. Update entity. Create new entity after.
```go
    existingUser.name = "new name"
    orm.Session(ctx).Flush()

    newUser := &user{name: "new user"}
    rep.Persists(ctx, newUser)
    orm.Session(ctx).Flush()
```

### Connect to database

D3 use wrappers on existing databases drivers. Currently, available pgx driver for postgresql.
This example of creating driver:

```go
	pgxConnection, _ := pgx.Connect(context.Background(), os.Getenv("DB"))
	d3Driver := d3pgx.NewPgxDriver(pgxConnection, &adapter.SquirrelAdapter{})
```

### Schema generation

You can generate database schema for registering entities. Use orm.GenerateSchema method instead:
```go
    // create connection and orm
    pgxConnection, _ := pgx.Connect(context.Background(), os.Getenv("DB"))
	d3orm := orm.NewOrm(d3pgx.NewPgxDriver(pgxConnection, &adapter.SquirrelAdapter{}))

    // register entities
    _ = d3orm.Register(&entity1{}, &entity2{})
	// generate schema sql - DDL commands
    sql, _ := d3orm.GenerateSchema()
    
    // exec schema sql for creating tables and indexes
	_, _ = pgxConnection.Exec(context.Background(), sql)
```

### Code generation

D3 using code generation for extract entity metadata and get access 
to entity fields. D3 require this code for every registered entity and
will be panic if don't find this code (panic will be throw in orm.Register method).
For generate code just run d3 tool with a file or directory where entities exist.
All entities must be comment with "//d3:entity".

### Session

### CRUD

### Repositories

### Relations

### Transactions

### First level cache

## Tests

## Roadmap

Current project status - is pre-alpha. It can be used in production with some risky.