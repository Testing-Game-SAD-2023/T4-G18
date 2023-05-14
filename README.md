# Game-Repository  
Implemented endpoint are:
* /games:
    - GET /{id} retrieve a game by ID;
    - GET /?pageSize=<integer>&page=<integer>&startDate=<YYYY-MM-DD>&endDate=<YYYY-MM-DD> list games in interval with pagination;
    - POST / create a game;
    - DELETE /{id} delete a game;
    - PUT /{id} update an existing game;
* /rounds:
    - GET /{id} retrieve a round by ID;
    - GET ?gameId=<integer> retrieve all rounds in a game;
    - POST / create a round;
    - DELETE /{id} delete a round;
    - PUT /{id} update an existing round;
* /turns:
    - GET /{id} retrieve a turn by ID;
    - GET ?roundId=<integer> retrieve all turns in a round;
    - GET /{id}/files retrieve player files as a zip;
    - PUT /{id}/files upload player files as a zip;
    - POST / create a turn;
    - DELETE /{id} delete a turn;
* /robots:
    - GET /?difficulty=<string>&type=<integer>&testClassId=<string> get a test result;
    - POST / create test result in bulk;
    - DELETE /?testClassId=<string> delete test result associated with a class; 

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
A swagger UI can be enabled setting `enableSwagger` to True in configuration file. Also, OpenAPI3 specification and a Postman collection are available in [`postman`](/postman) directory. When executing locally, SwaggerUI is available at http://localhost:3000/docs (if you change listenAddress in configuration, you should adapt the link to it).

### Testing
Unit testing is provided using mocking of storage class in order to not depend of a real database. Tests can be 
executed with:
```sh
make test
```
