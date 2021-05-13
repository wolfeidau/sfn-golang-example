# sfn-golang-example

This is a basic example of how to use [AWS step functions](https://aws.amazon.com/step-functions/) with lambda's written in [Go](https://golang.org).

# overview

The main things to call out features wise are:

* Uses [lambda middleware](https://github.com/wolfeidau/lambda-go-extras) to setup a logger in the context, and log all input and output events while in dev.
* Has an example of [error handling](https://docs.aws.amazon.com/step-functions/latest/dg/concepts-error-handling.html) using `Catch` feature of step functions.
* Illustrates how you can pass the [State Name in the input, and wrap parameters](sam/backend/sfn.yaml#L128-L129).

# License

This example is released under the Apache License, Version 2.0.