# Game-Repository - Prototype
In this repository a REST API skeleton is layed out. 
The prototype exposes basic endpoint for:
* games:
    - GET /games/{id} retrieve a game by ID;
    - POST /games create a game;
    - DELETE /games/{id}delete a game;
* rounds:
    - GET /rounds/{id} retrieve a round by ID;
    - POST /rounds create a round;
    - DELETE /rounds/{id} delete a round;

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

### Testing
As an example, a basic test suite (with a manual created mocked service for storage) is provided for the `GameController` struct.
Test can be executed using:
```
make test
```
Which includes the execution of all tests provided also the execution of race detector.

