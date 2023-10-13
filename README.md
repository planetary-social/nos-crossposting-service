# Nos Crossposting Service

## Building and running

Build the program like so:

    $ go build -o crossposting-service ./cmd/crossposting-service
    $ ./crossposting-service

The program takes no arguments. There is a Dockerfile available.

## Configuration

Configuration is performed using environment variables. This is also the case
for the Dockerfile.

### `CROSSPOSTING_LISTEN_ADDRESS`

Listen address for the main webserver in the format accepted by the Go standard
library.

Optional, defaults to `:8008` if empty.

### `CROSSPOSTING_METRICS_LISTEN_ADDRESS`

Listen address for the prometheus metrics server in the format accepted by the
Go standard library. The metrics are exposed under path `/metrics`.

Optional, defaults to `:8009` if empty.

### `CROSSPOSTING_ENVIRONMENT`

Execution environment. Setting environment to `DEVELOPMENT`:
- replaces an adapter which posts to Twitter with a noop adapter

Optional, can be set to `PRODUCTION` or `DEVELOPMENT`. Defaults to `PRODUCTION`.

### `CROSSPOSTING_LOG_LEVEL`

Log level.

Optional, can be set to `TRACE`, `DEBUG`, `ERROR` or `DISABLED`. Defaults to
`DEBUG`.

### `CROSSPOSTING_TWITTER_KEY`

Twitter API consumer key.

Required.

### `CROSSPOSTING_TWITTER_KEY_SECRET`

Twitter API consumer key secret.

Required.

### `DATABASE_PATH`

Full path to the database file.

Required, e.g. `/some/directory/database.sqlite`.

## Metrics

See configuration for the address of our metrics endpoint. Many out-of-the-box
Go-related metrics are available. We also have custom metrics:

- `application_handler_calls_total`
- `application_handler_calls_duration`
- `subscription_queue_length`
- `version`
- `public_key_downloader_count`
- `public_key_downloader_relays_count`
- `relay_connection_state`
- `twitter_api_calls_to_post_tweet`

See `service/adapters/prometheus`.

## Contributing

### Go version

The project usually uses the latest Go version as declared by the `go.mod` file.
You may not be able to build it using older compilers.

### How to do local development

#### Obtaining Twitter API keys

See ["How to get access to the Twitter API"][get-twitter-api-keys].

#### Run the service

Run the following command changing appropriate environment variables:

```
CROSSPOSTING_TWITTER_KEY=xxx \
CROSSPOSTING_TWITTER_KEY_SECRET=xxx \
CROSSPOSTING_DATABASE_PATH=/path/to/database.sqlite \
CROSSPOSTING_ENVIRONMENT=DEVELOPMENT \
go run ./cmd/crossposting-service
```

#### Updating frontend files

In order to update the embedded compiled frontend files run the following
command:

    $ make frontend

### Makefile

We recommend reading the `Makefile` to discover some targets which you can
execute. It can be used as a shortcut to run various useful commands.

You may have to run the following command to install a linter and a code
formatter before executing certain targets:

    $ make tools

If you want to check if the pipeline will pass for your commit it should be
enough to run the following command:

    $ make ci

It is also useful to often run just the tests during development:

    $ make test

Easily format your code with the following command:

    $ make fmt

### Writing code

Resources which are in my opinion informative and good to read:

- [Effective Go][effective-go]
- [Go Code Review Comments][code-review-comments]
- [Uber Go Style Guide][uber-style-guide]

#### Naming tests

When naming tests which tests a specific behaviour it is recommended to follow a
pattern `TestNameOfType_ExpectedBehaviour`. Example:
`TestRelayDownloader_EventsDownloadedFromRelaysArePublishedUsingPublisher`
.

#### Panicking constructors

Some constructors are prefixed with the word `Must`. Those constructors panic
and should always be accompanied by a normal constructor which isn't prefixed
with the `Must` and returns an error. The panicking constructors should only be
used in the following cases:
- when writing tests
- when a static value has to be created e.g. `MustNewHops(1)` and this branch of
  logic in the code is covered by tests

[get-twitter-api-keys]: https://developer.twitter.com/en/docs/twitter-api/getting-started/getting-access-to-the-twitter-api

[effective-go]: http://golang.org/doc/effective_go.html
[code-review-comments]: https://github.com/golang/go/wiki/CodeReviewComments
[uber-style-guide]: https://github.com/uber-go/guide/blob/master/style.md

