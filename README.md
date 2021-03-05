# obscurer
A compact library for equipping HTTP APIs with URL obscuring capabilities.

[![GoDoc][doc-img]][doc]
[![Build Status][ci-img]][ci]
[![Coverage Status][coverage-img]][coverage]
[![License][license-img]][license]

## What is it?

`obscurer` gives your HTTP API the ability to utilize obscure URLs, which
provides additional abstraction that is beneficial when implementing
level 3 RESTful APIs.

### What are obscure URLs?

Typically when creating a RESTful API using HTTP, the API defines URLs that
provide a predictable structure. For example, an API with a `product` resource
might have `/api/products/1/` URL path, where `1` is the ID of the product
being interacted with.

When creating [level 3][level-3-apis] hypermedia APIs, the API is responsible
for defining and managing the state of an application through links. As such,
any time a consumer is able to bypass this responsibility by issuing a
request to a URL that it wasn't given in the response of a previous request,
it actually undermines what a hypermedia API represents.

To prevent URLs from being "hackable" (or guessable), we obscure them. A
common method involves taking the normal URL path and running it through
a simple hashing algorithm, such as [MD5][md5]. These obscured URLs are then
returned to consumers in response representations, that they then use to issue
additional downstream requests.

## Why use it?

- can be extended to support any store of your choice (Redis, MySQL, etc.).
- immediately compatible with any request multiplexer.
- increased privacy of API interfaces.
- reduced "hackability" of API interfaces, promoting loose coupling.
- side-by-side support for unobscured and obscured URLs.
- automatically discards obscured URLs resulting in HTTP 404.

## How do I use it?

### Quickstart

```go
// create your mux.
mux := http.NewServeMux()
mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("i'm mando!")
})

// choose your store. 🎉
store := obscurer.DefaultStore

// choose your obscurer. 🎉
obscurer := obscurer.Default

// add obscured URL support. 🎉
handler := obscurer.NewHandler(obscurer, store, mux)

// create your server.
server := &http.Server{
	Addr: ":8080",
	Handler: handler, // 🎉
}

// start your server!
log.Fatal(server.ListenAndServe())
```
## Contribute

Want to lend us a hand? Check out our guidelines for
[contributing][contributing].

## License

We are rocking an [Apache 2.0 license][apache-license] for this project.

## Code of Conduct

Please check out our [code of conduct][code-of-conduct] to get up to speed
how we do things.

[level-3-apis]: https://www.crummy.com/writing/speaking/2008-QCon/act3.html
[md5]: https://en.wikipedia.org/wiki/MD5
[contributing]: https://github.com/freerware/obscurer/blob/main/CONTRIBUTING.md
[apache-license]: https://github.com/freerware/obscurer/blob/main/LICENSE.txt
[code-of-conduct]: https://github.com/freerware/obscurer/blob/main/CODE_OF_CONDUCT.md
[doc-img]: https://pkg.go.dev/badge/github.com/freerware/obscurer.svg
[doc]: https://pkg.go.dev/github.com/freerware/obscurer
[ci-img]: https://travis-ci.org/freerware/obscurer.svg?branch=main
[ci]: https://travis-ci.org/freerware/obscurer
[coverage-img]: https://codecov.io/gh/freerware/obscurer/branch/main/graph/badge.svg?token=yukEnN2Q2R
[coverage]: https://codecov.io/gh/freerware/obscurer
[license]: https://opensource.org/licenses/Apache-2.0
[license-img]: https://img.shields.io/badge/License-Apache%202.0-blue.svg
