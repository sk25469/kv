# KV - A Distributed In-Memory DB in Go

## Overview
This project is a key-value store implemented in Go, designed to provide a robust and scalable solution for storing and retrieving data efficiently. It supports a variety of features including Time to Live (TTL) for keys, transactions, publish-subscribe mechanisms, authentication, configuration file loading, replication, and sharding.


## Features
* **TTL for Keys**: Automatically expire and remove keys after a specified duration.

* **Transactions**: Group multiple operations in a single, atomic action.
Pub-Sub: Implement a publisher-subscriber model for message passing.


* **Authentication**: Secure access to the key-value store.

* **Configuration File Loading**: Easily configure the service through a JSON file.


* **Replication**: Ensure data durability and high availability through data replication.


* **Sharding**: Distribute data across multiple nodes to improve scalability and performance. (In progress)


## Setup Procedure

### Prerequisites

**Go 1.21.1 or higher**

### Installation

1. Clone the repository: 

```
git clone https://github.com/sk25469/kv.git
```

2. Build the server:

```
 go build -o kv-server
 ```

### Running the Server

Execute the built binary to start the server:

```
./kv-server
```


## Configuration
The server's behavior can be customized through a JSON configuration file. The default path for this file is specified in the server's main code. Ensure that the configuration file is correctly placed or update the path accordingly in the ``main.go`` file.


## Contributing
Contributions to the project are welcome. Please follow the standard fork and pull request workflow.


## License
This project is licensed under the MIT License. See the LICENSE file for details.


## Acknowledgments
This project uses several open-source libraries, as listed in go.mod and client/go.mod.


For more detailed information about the implementation and usage, refer to the source code documentation and comments.