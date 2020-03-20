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
  channels:
    - #ubuntu
    - #chat
    - #privgroup secret
    - #meeting keyring meetingchan
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

# Command-line

Planned examples:

1. `gic [ -f config ]`

Simply starts a gic instance.  Over stdin or the input socket you can
get a simple readline based interaction, i.e. "c #ubuntu" will choose
only channel #ubuntu for output, "t" will show a tail view, etc.

All configured forwarders will be activated.

1. `gic [ -f config ] -c #channel [ -c #channel ... ]`

Show a two-pane-per-channel curses view, with a 3-line pane at bootom
for input and the rest for channel output.  When channel starts with
`#` it is a channel, when it starts with `@` it is a special name,
so far pm or hilight, and otherwise it is a username.

1. `gic [ -f config ] -t [ chanlist ] [ -t [ chanlist ] ]`

Show a tail view of the listed chanlists.  A chanlist is a comma
separated list of channels and names.  Each `-t` entry results in
a multitail pane.  So `-t #ubuntu* -t myfriend -t @hilight` will show three
tail panes, one for all channels beginning with 'ubuntu', one for pms
with myfriend, and one for all hilights (pms and mentions).

1. `gic [ -f config ] grep [ -c chanlist ] [ -S startdate ] [ -U untildate ] regexp`

Grep the logs.
