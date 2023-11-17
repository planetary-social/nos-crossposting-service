# Nos Crossposting Service

A service which grabs nostr notes and posts them to Twitter as tweets. Right now
it only supports notes that aren't replies. The user opens the website, logs in
with their Twitter account and sets a list of npubs from which notes will be
cross-posted to their Twitter account.

## Design

### Retrieving relays

The idea behind the service is that it is pretty simple and easy to use. Because
of this we just ask the users to paste their npub in and ignore the matter of
their relays completely. Instead, we try to grab their relays from [Purple
Pages][purplepages] in the background. As it stands now it would seem that we
can't grab relay lists for ~50% of feeds. This is a problem that will hopefully
be rectified in the future in some way, perhaps by creating a service similar to
Purple Pages that crawls relays more aggressively.

### Twitter API errors

Posting tweets via the Twitter API seems to be failing often. We mostly get two
kinds of errors:
- rate limit exceeded
- unauthorized

I am not sure where the first one comes from since we are not exceeding rate
limits but perhaps it is somehow related to personal per-account limits of
Twitter accounts we are trying to post to. Unfortunately the API documentation
isn't helpful and doesn't explain this. The second one probably comes from users
revoking access for the app used by this service in their account settings.

If a tweet is stuck in the queue and can't be posted for a certain amount of
time we give up on posting it after several days and simply drop it completely.
The reasoning is that suddenly seeing tweets for notes that are quite old would
be confusing and probably not desired. Additionally there is almost no chance
that we will ever manage to post those tweets based on the metrics I am seeing.

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
- replaces a Twitter API adapter with a fake adapter
  - it doesn't actually post to Twitter
  - it returns hardcoded fake Twitter account details (due to weird rate-limiting errors)

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

### `CROSSPOSTING_DATABASE_PATH`

Full path to the database file.

Required, e.g. `/some/directory/database.sqlite`.

### `CROSSPOSTING_PUBLIC_FACING_ADDRESS`

Public facing address of the service, required for Twitter callbacks.

Required, e.g. `http://localhost:8008/` or `https://example.com/`.

## Obtaining Twitter API keys

The keys you are after are "Consumer keys". See ["How to get access to the
Twitter API"][get-twitter-api-keys].

Requirements:
- Your app must be a part of the project, it can't be standalone.
- App permissions need to be set to "Read and write".
- Type of app needs to be set to "Web App, Automated App or Bot".
- You need to set "Callback URI" accordingly:
  - for local development to e.g. `http://localhost:8008/login-callback`
    (unless you set `CROSSPOSTING_ENVIRONMENT` to `DEVELOPMENT` deactivating
    interacting with the Twitter API)
  - in production this has to be e.g. `https://example.com/login-callback`

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
- `twitter_api_calls`
- `accounts_count`
- `linked_public_keys_count`

See `service/adapters/prometheus`.

## Contributing

### Go version

The project usually uses the latest Go version as declared by the `go.mod` file.
You may not be able to build it using older compilers.

### How to do local development

Run the following command changing appropriate environment variables:

```
CROSSPOSTING_TWITTER_KEY=xxx \
CROSSPOSTING_TWITTER_KEY_SECRET=xxx \
CROSSPOSTING_DATABASE_PATH=/path/to/database.sqlite \
CROSSPOSTING_ENVIRONMENT=DEVELOPMENT \
CROSSPOSTING_PUBLIC_FACING_ADDRESS=http://localhost:8008/ \
go run ./cmd/crossposting-service
```

### Updating frontend files

Frontend is written in Vue and located in `./frontend`. Precompiled files are
supposed to be commited as they are embedded in executable files.

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


[purplepages]: https://purplepag.es/what