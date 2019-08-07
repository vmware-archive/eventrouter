# rockset-go-client
Official Go client library for Rockset

## Installation

Install the Rockset Go client from Github:

```
go get github.com/rockset/rockset-go-client
```

or install it from a source code checkout:

```
cd $GOPATH/src/github.com
mkdir rockset
cd rockset
git clone git@github.com:rockset/rockset-go-client.git
go install rockset-go-client/rockclient.go
```

## Usage

You can see a few [sample examples](https://github.com/rockset/rockset-go-client/tree/master/examples) of how to create a collection, how to put documents in a collection and how to use SQL to query your collections.

## Testing

Tests are available in the [Test](https://github.com/rockset/rockset-go-client/tree/master/test) folder.

Set ROCKSET_APIKEY and ROCKSET_APISERVER endpoint in the environment variables. To run tests:
```
go test ./test
```

## Support

Feel free to log issues against this client through GitHub.

## License

The Rockset Go Client is licensed under the [Apache 2.0 License](https://github.com/rockset/rockset-go-client/blob/master/LICENSE)
