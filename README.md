# JournalFields
This utility pretty prints journald messages with extra fields inline.
It uses [logrus][logrus] to reprint the log entries with the extra fields
visible inline.

This is specifically useful for projects that log extra fields using
[logrus][logrus] and the [journalhook][journalhook].
The extra fields feature in logrus is extremely useful, but for some reason
journalctl likes to hide those extra fields and make it comically difficult
to view the fields with the messages they are attached to.

# Example Usage

## The most basic usage.
```bash
journalctl -o json --user | journalfield
```

## Using a wrapper script

You could make the following wrapper script for journalctl and jounalfields.
The following script will pass all arguments to journalctl, enable json
output mode and, and pipe all output to journalfields.

```bash
#!/bin/bash

exec journalctl -o json "$@" | journalfield
```

[logrus]: https://github.com/sirupsen/logrus
[journalhook]: https://github.com/wercker/journalhook


# Developer Notes
I originally wanted to just call `journalctl` from within the main program.
The problem I ran into was that `journalctl` seemed like it was
checking its own calling name and trying to match it with some known
program names.
It was erroring out with some "Failed to match name blahhhh".