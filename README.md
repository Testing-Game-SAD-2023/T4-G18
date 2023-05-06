# Game-Repository  
Implemented endpoint are:
* games:
    - GET /games/{id} retrieve a game by ID;
    - POST /games create a game;
    - DELETE /games/{id} delete a game;
    - PUT /games/{id} update an existing game;
* rounds:
    - GET /rounds/{id} retrieve a round by ID;
    - POST /rounds create a round;
    - DELETE /rounds/{id} delete a round;
* turns:
    - GET /turns/{id} retrieve a turn by ID;
    - POST /turns create a turn;
    - DELETE /turns/{id} delete a turn;
    - PUT /turns/{id}/files upload player files as a zip;
    - GET /turns/{id}/files retrieve player files as a zip;

## Usage
The application uses a json configuration file like `config.example.json` (which contains default values). The application looks for a file `config.json` near executable, but a custom one can be provided through `--config=<PATH>` command line arguments.

The application can be executed through:
```sh
make run
```

Extra commands can be discovered using
```sh
make help
```

### Documentation
A swagger UI can be enabled setting `enableSwagger` to True in configuration file. Also, OpenAPI3 specification and a Postman collection are available in [`postman`](/postman) directory.

### Testing
Unit testing is provided using mocking of storage class in order to not depend of a real database. Tests can be 
executed with:
```sh
make test
```
