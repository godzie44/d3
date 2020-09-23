![golangci-lint](https://github.com/godzie44/d3orm/workflows/golangci-lint/badge.svg) ![tests](https://github.com/godzie44/d3orm/workflows/tests/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/godzie44/d3/badge.svg?branch=master)](https://coveralls.io/github/godzie44/d3?branch=master)

# D3 ORM

D3 is a golang ORM and DataMapper. This project was designed with respect to such 
ORM's like hibernate and doctrine. The main task - give an instrument to create a nice domain layer. If your business code not expressive enough, or if you want unit tests without a database, or if in your application, for some reason, you need aggregates, 
repositories, entities, and value objects - d3 can be a good choice.

## Motivation. Why another ORM?

In my opinion in GO have a lot of good ORM's. They are pretty fast and may save the developer from a lot of boilerplate code. But, if you want to write code
in DDD style (using DDD patterns and philosophy) it's not enough, 
because DDD approach imposes certain requirements. Main requirement - 
persistence ignorance. Current GO ORM's do not provide a sufficient level of abstraction for this.
In other words, we need to keep business logic free of data access code. That's why D3 created.

## Main futures

- code generation instead of reflection
- DB schema auto-generation
- one-to-one, one-to-many, and many-to-many relations
- lazy and eager relation loading
- query builder
- relation fetch strategies (eager/lazy as above or extract relation in one query with join)
- fetched entity cache (first level cache)
- cascade remove and update of related entities
- application-level transaction (UnitOfWork)
- DB transactions support
- smart persist layer doesn't generate redundant queries on entity updates
- UUID support

## Documentation

All documentation is on the project [wiki](https://github.com/godzie44/d3/wiki).

## Example of usage

A full example of usage you can see in [this](https://github.com/godzie44/last-wish) project. Note, that is it is a test project
with main responsibility show d3 ORM in action.

## Tests

D3 integration tests require a PostgreSQL database. Connect to the database specified in the D3_PG_TEST_DB environment variable.

## Roadmap

- [ ] add mysql support
- [X] add sqlite support
- [ ] composite pk
- [ ] not only schema generation but generation of schema diff's
- [ ] embedding structures
- [X] index definition in entity comments
- [ ] generate fk's for relations

Note: Current project status - is alpha. It can be used in production with some risky.
