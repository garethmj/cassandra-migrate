# Cassandra Migrate

Yet another "migrations" tool for Cassandra. This came about because, having tried out and evaluated several, non of them
seemed to do quite what I wanted. One of those things is to have support for having scripts that will *only* run in a
particular environment. There are some other reasons...which I forget.

## Usage

Don't. No, really, I wouldn't. This is still a work in progress and, whilst it seems to be working just fine for what I
want, I'd not like to suggest that it will be generically applicable at this time. A number of assumptions are made that
fit our specific use case but might not be what others are looking for.

I'll try to make the tool better and more generic in coming weeks.

## Road Map

Things I'd like to add:

* A proper CQL lexer and parser. I can't find one in Go so I guess I'd have to write one? This would really help in a number of ways.
* Migrate down. At the moment we only go forwards. Very progressive. But not always what you want.
* In-file meta-data. Something that a parser would be really helpful for. But being able to add meaningful annotation to a CQL file would be ace.
* Validate checksums: We sha1sum all the files and add that info to the schema_version table but never audit it.
* Stop fmt.Printf'ing and use a logger instead.
* Manage dependencies.
