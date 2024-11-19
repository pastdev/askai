# Ask AI

A tool for interacting with an AI service from the command line.

## Using Ollama

Start ollama on windows, need to specify 0.0.0.0 or it will only listen on 127.0.0.1

```console
PS C:\Users\lucas> $env:OLLAMA_HOST="0.0.0.0:11434"
PS C:\Users\lucas> Get-NetTCPConnection -LocalPort 11434

LocalAddress                        LocalPort RemoteAddress                       RemotePort State       AppliedSetting OwningProcess
------------                        --------- -------------                       ---------- -----       -------------- -------------
127.0.0.1                           11434     0.0.0.0                             0          Listen                     10276


PS C:\Users\lucas> ollama serve
```

Then locate the IP address and associated with this service from WSL:

```console
ltheisen@ltserver ~/git/pastdev/askai
$ grep ^nameserver /etc/resolv.conf
nameserver 172.22.144.1
```

This information will be needed for your [configuration](#configuration).

## Configuration

Configuration is loaded by default from the following directories (in order):

* /etc/askai.d
* ~/.config/askai.d
* ./askai.d

The format of configuration is:

```yaml
default_endpoint: windows_ollama
endpoints:
  windows_ollama:
    api_type: OPEN_AI
    base_url:            "http://172.22.144.1:11434/v1"
    empty_messages_limit: 300
```

More options are available, see the `Config` type in the `askai` package for details.
