# Tsdecls

This is tsdecls,
a command and library for parsing the Go implementation of a webapp server
and producing Typescript declarations for the client side.

## Installation

If you want to use the `tsdecls` command:

```sh
go install github.com/bobg/tsdecls/cmd/tsdecls@latest
```

If you want to use the `tsdecls.Write` function from within your Go program:

```sh
go get github.com/bobg/tsdecls
```

## Description

Web-based applications are divided into a server side and a client side.
The server side stores and manipulates persistent state.
The client side contains the user interface.
User interaction with the interface
causes it to send requests to the server side.

Normally this means that the application writer must implement parallel functionality on both sides:
marshaling requests for network transit on the client side,
and unmarshaling responses;
and on the server side,
unmarshaling requests and marshaling responses.
Errors can arise if the server-side and client-side implementations do not agree.

Tsdecls simplifies this by producing the client-side logic _from_ the server-side logic.
Given a directory containing a Go package,
and the name of a type in that package,
tsdecls produces Typescript declarations for the eligible methods in that type.

A method is eligible if it is exported
and has the signature `func(context.Context, X) (Y, error)`
for some JSON-marshalable types X and Y.
Each argument and result is optional;
e.g. a method is still eligible if its signature is `func(X) error`.

A method with this signature can be turned into an HTTP handler
using [mid.JSON](https://pkg.go.dev/github.com/bobg/mid#JSON)
from the github.com/bobg/mid library.

## Output

Sample Typescript output can be seen at [github.com/bobg/tsdecls/blob/master/_testdata/want.ts](https://github.com/bobg/tsdecls/blob/master/_testdata/want.ts).

The output begins with the definition of a function, `post`,
for JSON-encoding the (optional) request body,
sending a request to a given endpoint on the server
(presumed to be at the same host as in `window.location`),
and JSON-decoding the (optional) response.

There follows a number of `interface` definitions for the types used in the various methods.
This list of definitions is topologically sorted;
i.e., if A is used in the definition of B,
then the definition of A will appear first.

Finally, there is an object with the same name as the Go type.
Each eligible Go method `FooBar` is turned into a Typescript method `foo_bar`
(i.e., “CamelCase” is converted to “snake_case”),
which takes an argument of the appropriate request type (if any),
calls `post` with the endpoint `/s/foo_bar` and the request argument,
and casts the result of that to a `Promise` of the response type (if any).
