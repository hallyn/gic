# gic: go irc client

This is a simple irc client.

# Features

1. config file driven connections
1. ssl support
1. password from kernel keyring
1. connection output to a single file
1. flexible regexp based output to
   1. files/pipes
   1. shell-out programs
1. simple input api over pipe

# Example config:

```
server:
  host: irc.freenode.net
  port: 6697
  ssl: true
  password: keyring irc
  nick: goonie
config:
  input:
    type: named [ || abstract ]
    path: $HOME/.irc-in
  output: default
```

# Input

Input by default will be over standard input.  If the config file
lists an alternate input location, it will interpreted as a
unix socket.  If 'input' is specifed as 'default', then it will
be a unix socket named `$HOME/.cache/gic/${server}/in`.

# Output

By default, all output goes to standard output.

If the ocnfig file lists a `output`, then under that path
will be created a series of files:

1. A file named 'server.out' will contain all output for the server
1. A file named hilights will contain all PMs lines and all channel
   lines where you were mentioned.
1. Under the 'channels/' directory, a file named for each channel will
   contain all output for that channel
1. Under the 'people/' directory, a file named for each PM'd person will
   contain all those PMs.

If `output` is set to "default", then the output path will be
`$HOME/.cache/gic/${server}`.

# Parsers

Parsers list rules and destination files or programs to flexibly handle
irc events.  Rules include:

1. pm [ all | first | username ]
1. mention
1. regexp (all lines matching the listed regexp)

Destinations can be specific named files, filenames built from the rule name,
or programs to shell out to, which will receive the triggering line as
standard input.
