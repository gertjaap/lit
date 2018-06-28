# lit - a lightning node you can run on your own
![Lit Logo](litlogo145.png)

[![Build Status](http://hubris.media.mit.edu:8080/job/lit-PR/badge/icon)](http://hubris.media.mit.edu:8080/job/lit-PR/)

Under development, not for use with real money.

## Setup

### Prerequisites
- [Git](https://git-scm.com/)
- [Go](https://golang.org/doc/install)

### Installing

1. Clone this repo with `git clone https://github.com/mit-dci/lit` or do `go get github.com/mit-dci/lit`

2. `cd` into the `lit` directory (either inside your GOPATH or your cloned directory)

3. Run `make lit` to build lit and `make test` to run the tests. `make test with-python=true` will include the python tests (requires `bitcoind`). Alternatively, you can run `go build` to build lit if you're building inside your GOPATH.

4. Run `./lit --tn3 1` to start lit

The words `yup, yes, y, true, 1, ok, enable, on` can be used to specify that lit automatically connect to a set of populated seeds. It can also be replaced by the ip of the remote node you wish to connect to.

## Using Lightning

Great! Now that you are all done setting up lit, you can
- read about the arguments for starting lit [here](#command-line-arguments)
- read about the folders for the code and what does what [here](#folders)
- head over to the [Walkthrough](./WALKTHROUGH.md) to create some lit nodes or
- check out how to [Contribute](./CONTRIBUTING.md).

## Command line arguments

When starting lit, the following command line arguments are available. The following commands may also be specified in lit.conf which is automatically generated on startup.

#### connecting to networks:

| Arguments                   | Details                                                      | Default Port  |
| --------------------------- |--------------------------------------------------------------| ------------- |
| `--tn3 <nodeHostName>`      | connect to `nodeHostName`, which is a bitcoin testnet3 node. | 18333         |
| `--reg <nodeHostName>`      | connect to `nodeHostName`, which is a bitcoin regtest node.  | 18444         |
| `--lt4 <nodeHostName>`      | connect to `nodeHostName`, which is a litecoin testnet4 node.| 19335         |

#### other settings:

| Arguments                   | Details                                                      |
| --------------------------- |--------------------------------------------------------------|
| `-v` or `--verbose`         | Verbose; log everything to stdout as well as the lit.log file.  Lots of text.|
| `--dir <folderPath>`        | use `folderPath` as the directory.  By default, saves to `~/.lit/` |
| `-p` or `--rpcport <portNumber>` | listen for RPC clients on port `portNumber`.  Defaults to `8001`.  Useful when you want to run multiple lit nodes on the same computer (also need the `--dir` option) |
| `-r` or `--reSync`          | try to re-sync to the blockchain |

## Folders

| Folder Name  | Details                                                                                                                                  |
|:-------------|:-----------------------------------------------------------------------------------------------------------------------------------------|
| `cmd`        | Has some rpc client code to interact with the lit node.  Not much there yet                                                              |
| `elkrem`     | A hash-tree for storing `log(n)` items instead of n                                                                                      |
| `litbamf`    | Lightning Network Browser Actuated Multi-Functionality -- web gui for lit                                                                |
| `litrpc`     | Websocket based RPC connection                                                                                                           |
| `lndc`       | Lightning network data connection -- send encrypted / authenticated messages between nodes                                               |
| `lnutil`     | Some widely used utility functions                                                                                                       |
| `portxo`     | Portable utxo format, exchangable between node and base wallet (or between wallets).  Should make this into a BIP once it's more stable. |
| `powless`    | Introduces a web API chainhook in addition to the uspv one                                                                               |
| `qln`        | A quick channel implementation with databases.  Doesn't do multihop yet.                                                                 |
| `sig64`      | Library to make signatures 64 bytes instead of 71 or 72 or something                                                                     |
| `test`       | Integration tests                                                                                                                        |
| `uspv`       | Deals with the network layer, sending network messages and filtering what to hand over to `wallit`                                       |
| `wallit`     | Deals with storing and retrieving utxos, creating and signing transactions                                                               |
| `watchtower` | Unlinkable outsourcing of channel monitoring                                                                                             |

### Hierarchy of packages

One instance of lit has one litNode (package qln).

LitNodes manage lndc connections to other litnodes, manage all channels, rpc listener, and the ln.db.  Litnodes then initialize and contol wallits.

A litNode can have multiple wallits; each must have different params.  For example, there can be a testnet3 wallit, and a regtest wallit.  Eventually it might make sense to support a root key per wallit, but right now the litNode gives a rootPrivkey to each wallet on startup.  Wallits each have a db file which tracks utxos, addresses, and outpoints to watch for the upper litNode.  Wallits do not directly do any network communication.  Instead, wallits have one or more chainhooks; a chainhook is an interface that talks to the blockchain.

One package that implements the chainhook interface is uspv.  Uspv deals with headers, wire messages to fullnodes, filters, and all the other mess that is contemporary SPV.

(in theory it shouldn't be too hard to write a package that implements the chainhook interface and talks to some block explorer.  Maybe if you ran your own explorer and authed and stuff that'd be OK.)

#### Dependency graph

![Dependency Graph](deps-2018-06-19.png)

## License
[MIT](https://github.com/mit-dci/lit/blob/master/LICENSE)
