# Introduzione
Il componente realizza una API REST per gestire le informazioni che riguardano le partite. In particolare, è composta da:

* Rest server: componente principale dell'applicazione realizzato dal team
* Prometheus: permette l'estrazione delle metriche dal rest-server con un sistema di monitoraggio periodico;
* Grafana: elabora le metriche estratte da Prometheus per la creazione di dashboard grafiche per misurare le performance dell'applicazione;
* Postgres: database relazionale necessario al funzionanmento dell'applicazione;
Ad eccezione del rest-server e di postgres, tutti gli altri componenti sono opzionali.

## Dipendenze
In questa sezione, viene spiegato in dettaglio come è stato sviluppato il componente rest-server. Il rest-server è un'applicazione GO ed è dotata di un `Makefile` per gestire il processo di sviluppo e compilazione. 

### Installazione toolchain GO
Dato che Go è dotato di un runtime multipiattaforma, per installare la toolchain per il proprio sistema basta navigare sul sito ufficiale (https://go.dev/dl/), scaricare la versione adatta al proprio sistema e seguire le istruzioni. In particolare, la versione utilizzata per sviluppare il progetto è la `1.20.4`.

#### Dipendenze di progetto
Visto che l'applicazione usa il `vendoring` delle dipendenze non è necessario installare alcuna libreria.

### Configurazione
L'applicazione deve essere configurata con un file in formato `json` il cui path deve essere passato con l'argomento `--config=<PATH>`. Il comportamento default è quello di cercare un file `config.json` all'interno della directory corrente. I valori di default della configurazione sono riportati nel file `config.example.json`.

```json
{
    "postgresUrl": "",
    "listenAddress": "localhost:3000",
    "apiPrefix": "/",
    "dataPath": "data",
    "enableSwagger": false,
    "rateLimiting": {
        "enabled": false,
        "burst": 4,
        "maxRate": 2
    }
}

```
Prima di eseguire l'applicazione è necessario creare un file di configurazione funzionante. In particolare, tutti i parametri specificati sopra sono opzionale ad eccezione di postgresUrl. Quindi, assicurandosi di avere un'istanza Postgres funzionante, bisogna creare un file di configurazione `config.json` nella cartella principale del progetto con il seguente contentuto:


```json title="config.json"
{
    "postgresUrl": "<POSTGRES_URL>",
}
```

### Compilazione ed esecuzione
L'applicazione può essere compilata con:
```sh
make build
```

Mentre per eseguirla:

```sh
make run
```
Questo comando esegue l'applicazione cercando il file di configurazione nella cartella corrente. Per specificare un path custom:

Mentre per eseguirla:

```sh
make run CONFIG=<PATH_TO_CONFIG>
```
## Installazione
Ci sono diverse modalità per installare l'applicazione. In particolare, in questa sezione si farà riferimento solo al rest-server.

### Docker-from source
Una volta effettuato il `clone` della repository