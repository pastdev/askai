# Ask AI

A tool for interacting with an AI service from the command line.

~~~console
ltheisen@MM269697-PC ~
$ askai complete --user "how do i use logs in scriptrunner for jira" --max-tokens 1000 --stream
1. Import the `Logger` class:
   ```
   import com.onresolve.scriptrunner.runner.util.Logger
   ```

2. Create a logger instance:
   ```
   def log = Logger.log(this.class.name)
   ```

3. Log at different levels using the following methods:
   - `log.debug()`: for debug messages
   - `log.info()`: for informational messages
   - `log.warn()`: for warning messages
   - `log.error()`: for error messages

Example:
   ```
   log.debug('Debug message')
   log.info('Info message')
   log.warn('Warning message')
   log.error('Error message')
   ```

4. Output of the log messages is displayed in the ScriptRunner logs:
   Jira → Issues and filters → ScriptRunner → Logs
~~~

## Installation

`askai` is a self contained binary that has [pre-built releases for various platforms](https://github.com/pastdev/askai/releases).
You may find this script valuable for installation:

~~~bash
# note this command uses clconf which can be found here:
#   https://github.com/pastdev/clconf
(
  # where do you want this installed?
  binary="${HOME}/.local/bin/askai"
  # one of linux, darwin, windows
  platform="linux"
  curl \
    --location \
    --output "${binary}" \
    "$(
      curl --silent https://api.github.com/repos/pastdev/askai/releases/latest \
        | clconf \
          --pipe \
          jsonpath "$..assets[*][?(@.name =~ /askai-${platform/windows/windows.exe}/)].browser_download_url" \
          --first)"
  chmod 0755 "${binary}"
)
~~~

### The ai alias

You may also find it useful to add this alias to one of your shell profile scripts:

~~~bash
alias ai='askai complete --user '
~~~

## Configuration

Configuration is loaded by default from the following directories (in order):

* `/etc/askai.d`
* `~/.config/askai.d`
* `./askai.d`

The format of configuration is:

~~~yaml
default_endpoint: windows_ollama
endpoints:
  # An API endpoint with both chat completion and image generation models
  grok:
    api_type: OPEN_AI
    auth_token: <OMITTED>
    base_url: "https://api.x.ai/v1"
    # optionally supply defaults for requests
    chat_completion_defaults:
      messages:
      - content: |
          provide concise responses with minimal ceremony. do not use a
          conversational style.
        role: system
      model: grok-3-latest
    image_defaults:
      model: grok-2-image-latest
  windows_ollama:
    api_type: OPEN_AI
    base_url: "http://172.22.144.1:11434/v1"
    empty_messages_limit: 300
    # optionally supply defaults for requests
    chat_completion_defaults:
      max_tokens: 250
      messages:
      - content: |
          provide concise responses with minimal ceremony. do not use a
          conversational style.
        role: system
      model: mistral
~~~

More options are available, see the `Config` type in the `askai` package for details.

## Using Ollama

Start ollama on windows.
You need to specify `0.0.0.0` or it will only listen on `127.0.0.1`.

~~~console
PS C:\Users\lucas> $env:OLLAMA_HOST="0.0.0.0:11434"
PS C:\Users\lucas> Get-NetTCPConnection -LocalPort 11434

LocalAddress                        LocalPort RemoteAddress                       RemotePort State       AppliedSetting OwningProcess
------------                        --------- -------------                       ---------- -----       -------------- -------------
127.0.0.1                           11434     0.0.0.0                             0          Listen                     10276


PS C:\Users\lucas> ollama serve
~~~

Then locate the IP address and associated with this service from WSL:

~~~console
ltheisen@ltserver ~/git/pastdev/askai
$ grep ^nameserver /etc/resolv.conf
nameserver 172.22.144.1
~~~

This information will be needed for your [configuration](#configuration).

## Development

This project has a _snapshot_ script that makes using a snapshot build version of the source easy.
To make use of this feature, create a symlink to the script somewhere in your path.
For example:

~~~bash
# assuming you are currently in the base directory of this project
ln -s "$(pwd)/askais ~/.local/bin/askais
~~~

If you like [the `ai` alias](#the-ai-alias), you can updated it to point to this script:

~~~bash
alias ai="askais complete --user"
~~~
