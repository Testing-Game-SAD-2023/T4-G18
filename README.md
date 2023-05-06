# Game-Repository  
Implemented endpoint are:
* games:
    - GET /games/{id} retrieve a game by ID;
    - GET /games?pageSize=<integer>&page=<integer>&startDate=<YYYY-MM-DD>&endDate=<YYYY-MM-DD> retrieve a game by ID;
    - POST /games create a game;
    - DELETE /games/{id} delete a game;
    - PUT /games/{id} update an existing game;
* rounds:
    - GET /rounds/{id} retrieve a round by ID;
    - GET /rounds?gameId=<integer> retrieve all rounds in a game;
    - POST /rounds create a round;
    - DELETE /rounds/{id} delete a round;
    - PUT /rounds/{id} update an existing round;
* turns:
    - GET /turns/{id} retrieve a turn by ID;
    - GET /turns?roundId=<integer> retrieve all turns in a round;
    - GET /turns/{id}/files retrieve player files as a zip;
    - PUT /turns/{id}/files upload player files as a zip;
    - POST /turns create a turn;
    - DELETE /turns/{id} delete a turn;

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
