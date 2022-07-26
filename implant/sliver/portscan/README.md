# portscan.exe

A simple port scanner for use on EPT because scanning over SOCKS is awful.

```
Usage: portscan.exe -h hostspec -p portspec -t threads [default 8]

Examples:
  portscan.exe -h 10.0.0.0/24 -p 21-23,389,3389
  portscan.exe -h vcenter01.dbg.local -p 443
  portscan.exe -h hosts.txt -p 135,139,445 -t 16
```

```
> echo scanme.nmap.org > hosts.txt
> portscan.exe -h hosts.txt -p 20-23,80,443
scanme.nmap.org:80
scanme.nmap.org:22
```
```
> portscan.exe -h 10.5.0.0/24 -p 80,443,445
10.5.0.42:443
10.5.0.47:445
10.5.0.48:445
10.5.0.49:445
10.5.0.51:443
10.5.0.57:80
10.5.0.61:443
```
