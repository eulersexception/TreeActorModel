## Ausführen mit Docker

-   Images bauen

    ```
    make docker
    ```

-   ein (Docker)-Netzwerk `actors` erzeugen

    ```
    docker network create actors
    ```

-   Starten des Tree-Services und binden an den Port 8090 des Containers mit dem DNS-Namen
    `treeservice` (entspricht dem Argument von `--name`) im Netzwerk `actors`:

    ```
    docker run --rm --net actors --name treeservice treeservice \
      --bind="treeservice.actors:8090"
    ```

    Damit das funktioniert, müssen Sie folgendes erst im Tree-Service implementieren:

    -   die `main` verarbeitet Kommandozeilenflags und
    -   der Remote-Actor nutzt den Wert des Flags
    -   wenn Sie einen anderen Port als `8090` benutzen wollen,
        müssen Sie das auch im Dockerfile ändern (`EXPOSE...`)

-   Starten des Tree-CLI, Binden an `treecli.actors:8091` und nutzen des Services unter
    dem Namen und Port `treeservice.actors:8090`:

    ```
    docker run --rm --net actors --name treecli treecli --bind="treecli.actors:8091" \
      --remote="treeservice.actors:8090" trees
    ```

    Hier sind wieder die beiden Flags `--bind` und `--remote` beliebig gewählt und
    in der Datei `treeservice/main.go` implementiert. `trees` ist ein weiteres
    Kommandozeilenargument, dass z.B. eine Liste aller Tree-Ids anzeigen soll.

    Zum Ausprobieren können Sie den Service dann laufen lassen. Das CLI soll ja jedes
    Mal nur einen Befehl abarbeiten und wird dann neu gestartet.

-   Zum Beenden, killen Sie einfach den Tree-Service-Container mit `Ctrl-C` und löschen
    Sie das Netzwerk mit

    ```
    docker network rm actors
    ```

## Ausführen mit Docker ohne vorher die Docker-Images zu bauen

Nach einem Commit baut der Jenkins, wenn alles durch gelaufen ist, die beiden
Docker-Images. Sie können diese dann mit `docker pull` herunter laden. Schauen Sie für die
genaue Bezeichnung in die Consolenausgabe des Jenkins-Jobs.

Wenn Sie die Imagenamen oben (`treeservice` und `treecli`) durch die Namen aus der
Registry ersetzen, können Sie Ihre Lösung mit den selben Kommandos wie oben beschrieben,
ausprobieren.

## Ausführen mit Docker

-   Images bauen

    ```
    make docker
    ```

-   ein (Docker)-Netzwerk `actors` erzeugen

    ```
    docker network create actors
    ```

-   Starten des Tree-Services und binden an den Port 8090 des Containers mit dem DNS-Namen
    `treeservice` (entspricht dem Argument von `--name`) im Netzwerk `actors`:

    ```
    docker run --rm --net actors --name treeservice treeservice \
      --bind="treeservice.actors:8090"
    ```

    Damit das funktioniert, müssen Sie folgendes erst im Tree-Service implementieren:

    -   die `main` verarbeitet Kommandozeilenflags und
    -   der Remote-Actor nutzt den Wert des Flags
    -   wenn Sie einen anderen Port als `8090` benutzen wollen,
        müssen Sie das auch im Dockerfile ändern (`EXPOSE...`)

-   Starten des Tree-CLI, Binden an `treecli.actors:8091` und nutzen des Services unter
    dem Namen und Port `treeservice.actors:8090`:

    ```
    docker run --rm --net actors --name treecli treecli --bind="treecli.actors:8091" \
      --remote="treeservice.actors:8090" trees
    ```

    Hier sind wieder die beiden Flags `--bind` und `--remote` beliebig gewählt und
    in der Datei `treeservice/main.go` implementiert. `trees` ist ein weiteres
    Kommandozeilenargument, dass z.B. eine Liste aller Tree-Ids anzeigen soll.

    Zum Ausprobieren können Sie den Service dann laufen lassen. Das CLI soll ja jedes
    Mal nur einen Befehl abarbeiten und wird dann neu gestartet.

-   Zum Beenden, killen Sie einfach den Tree-Service-Container mit `Ctrl-C` und löschen
    Sie das Netzwerk mit

    ```
    docker network rm actors
    ```

## Ausführen mit Docker ohne vorher die Docker-Images zu bauen

Nach einem Commit baut der Jenkins, wenn alles durch gelaufen ist, die beiden
Docker-Images. Sie können diese dann mit `docker pull` herunter laden. Schauen Sie für die
genaue Bezeichnung in die Consolenausgabe des Jenkins-Jobs.

Wenn Sie die Imagenamen oben (`treeservice` und `treecli`) durch die Namen aus der
Registry ersetzen, können Sie Ihre Lösung mit den selben Kommandos wie oben beschrieben,
ausprobieren.

Um den Baum zu testen können einfach folgende Befehle eingegeben werden. Die fertigen Docker Images
für den TreeService und die TreeCli werden dann gepulled und die Anwendung kann per Eingaben über die
Kommandozeile genutzt werden.

### TreeService

TreeService starten:

```
docker run --rm --net actors --name treeservice terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treeservice \
--bind="treeservice.actors:8091"
```

Nach dem Start läuft der TreeService bis ihn der User mit der Tastenkombination ```Ctrl+c``` terminiert.

### TreeCli

####  In <Klammern> angegebene Ausdrücke in den Befehlen vor der Ausführung entsprechend einsetzen. 

* Baum erzeugen:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" newtree <int>
```

* Schlüssel-Wert-Paar in Baum einfügen:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" --id=<Tree-ID> --token=<String> insert <int> <String>
``` 

* Wert mit Schlüssel suchen:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" --id=<Tree-ID> --token=<String> search <int>
``` 

* Wert mit Schlüssel löschen:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" --id=<Tree-ID> --token=<String> delete <int>
``` 

* Alle Schlüssel-Wert-Paare eines Baums erhalten:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" --id=<Tree-ID> --token=<String> traverse
``` 

* Einen Baum löschen:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" --id=<Tree-ID> --token=<String> deletetree
``` 

* Die Anfrage, um einen Baum zu löschen wird zwar bestätigt, aber nicht final ausgeführt, damit der User vor einer versehentlichen Löschung gewarnt ist. Der folgende Befehl löscht den Baum endgültig:

```
docker run --rm --net actors --name treecli terraform.cs.hm.edu:5043/ob-vss-ws19-blatt-3-suedachse:PR-8-treecli --bind="treecli.actors:8090" --remote="treeservice.actors:8091" --id=<Tree-ID> --token=<String> forcetreedelete
``` 
