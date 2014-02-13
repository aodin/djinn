Djinn
=====

Djinn is a port of Django's user, session, and permission systems to Go. It is intended to integrate seamlessly with Go's `net/http` and `database/sql` packages.

It treats Django 1.6's default database models and session cookies as an API and is able to create, read, and update those interfaces.

Check the wiki for examples.

Caveats:

* Instead of pickling and un-pickling session data it is encoded by the Go `encoding/json` package. Pickled and JSON data are similar enough that the default session data will work. As of Django 1.6, session data will be encoded using JSON by default.

The D is silent.

2014