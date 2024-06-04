## StructMap Service

StructMap Service is a powerful and easy-to-use HTTP-based service for managing in-memory data structures. It leverages the `compactmap` and `structmap` libraries to provide efficient storage and retrieval of data. This repository contains the server implementation using FastHTTP and a client library to interact with the server.

### Features

- **Add**: Add new items to the storage.
- **Get**: Retrieve items by ID.
- **Delete**: Remove items by ID.
- **Update**: Update items based on conditions.
- **SetField**: Update a single field of an item.
- **SetFields**: Update multiple fields of an item.
- **Find**: Find items based on conditions.
- **Iterate**: Iterate over all items in the storage.
- **Clear**: Clear all items in the storage.

### Server

The server uses FastHTTP for high-performance HTTP handling. It exposes a set of API endpoints to interact with the in-memory data structures.

  **API Endpoints**:
    - `POST /api/add`
    - `GET /api/get?id=<id>`
    - `GET /api/delete?id=<id>`
    - `POST /api/update`
    - `POST /api/setfield`
    - `POST /api/setfields`
    - `POST /api/find`
    - `GET /api/iterate`
    - `GET /api/clear`

### Client

The client library provides convenient methods to interact with the StructMap Service via HTTP. It simplifies sending requests and handling responses from the server.

### Why Use StructMap Service?

StructMap Service offers a simple yet powerful way to manage in-memory data structures via HTTP API. This setup is particularly useful for:

- **Microservices**: Easily interact with in-memory data across different services.
- **Rapid Prototyping**: Quickly set up a backend service to handle data without setting up a database.
- **Scalability**: High