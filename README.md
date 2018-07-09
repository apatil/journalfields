# JournalFields
This utility pretty prints journald messages with extra
[fields][journald-fields] inline.
It uses [logrus][logrus] to reprint the log entries with the extra fields
visible inline.

This is specifically useful for projects that log extra fields using
[logrus][logrus] and the [journalhook][journalhook].
The extra fields feature in logrus is extremely useful, but for some reason
journalctl likes to hide those extra fields and make it comically difficult
to view the fields with the messages they are attached to.
Using native journald fields has the advantage that you can search across
multiple services for a field with certain values.

# Grab it
```bash
go get github.com/linux4life798/journalfields
```

# Example Usage

## The most basic usage
```bash
journalctl -o json --user | journalfields
```

## Basic usage with field selection
Say your service, called `lorawan`, logs with many fields, but you are only
interested in the `DEVEUI` and `APPID` fields.

```bash
journalctl -o json -u lorawan | journalfields DEVEUI APPID
```

## Using the wrapper script

I have included a wrapper script for `journalctl` and `jounalfields`,
called [journalctlf](journalctlf).

* It passes all arguments before a `--` to `journalctl`
* It passes all arguments after `--`, as fields, to `journalfields`
* It enables journalctl's json output mode and pipes all output to journalfields

To acomplish the same result as the previous use case, you could do the following:
```bash
journalctlf -u lorawan -- DEVEUI APPID
```

The following is from the [journalctlf](journalctlf) wrapper script:
```bash
#!/bin/bash
# This script calls journalctl with the json output format and pipes it
# to journalfields.
# * Arguments before '--' are passed to journalctl
# * Arguments after '--' are passed to journalfields (the selected fields)

# This should probably be represented are an absolute path, so that you
# can call thiis with sudo
JOURNALFIELDS=journalfields

JOURNALCTL_ARGS=( )
FIELDS=( )

# Split arguments given at the --
for arg; do
	shift
	if [ "$arg" = "--" ]; then
		break
	fi

	JOURNALCTL_ARGS+=( "$arg" )
done

FIELDS=( "$@" )

echo exec journalctl -o json "${JOURNALCTL_ARGS[@]}" \| $JOURNALFIELDS "${FIELDS[@]}"
exec journalctl -o json "${JOURNALCTL_ARGS[@]}" | $JOURNALFIELDS "${FIELDS[@]}"
```

## Filter by field using the wrapper script

Here is where the power of journald's fields really comes into play.

Say you want to show all log entries from that same `lorawan` service
that have the field `DEVEUI` set to `1122334455667788`.
Additionally, you want to show the same `DEVEUI` and `APPID` fields.

You could do the following:
```bash
journalctlf -u lorawan DEVEUI=1122334455667788 -- DEVEUI APPID
```

## Filter by field across multiple services

Journalctl even allows you to filter by a given set of fields across all
logged content (from all services).

Say you have a group of systemd services that all use the common field `DEVEUI`
in their logged messages.
By simply dropping the `-u lorawan` from the previous use case, you can
search for all log entries that have the `DEVEUI` set to `1122334455667788`.

You could do the following with the wrapper script:
```bash
journalctlf DEVEUI=1122334455667788 -- DEVEUI APPID
```

Or, without the wrapper script:
```bash
journalctl -o json DEVEUI=1122334455667788 | journalfields DEVEUI APPID
```

# Using this tool in your workflow

As I mentioned above, this tool really comes into action when you start using
the native journal logging interface(or syslog compatibility interface), where
optional fields can be given.
This aligns well Go's [logrus][logrus] library, where fields are prominent,
and very useful for debugging.
The only problem is that when you start using the logrus native
journald interface hook, [journalhook][journalhook], those pretty fields that
you could easily read before become hidden in journalctl.
I have yet to find the right combination of arguments to have journalctl
spit out those precious fields on the same line as the message.
Since people, including myself, really like the output format of
[logrus][logrus], I have decided to make a tool that is capable of
re-interpreting the journald logs as logrus Entries and printing them out in
colorful logrus style.

JournalFields gives you the feeling that the log messages were printed
directly from logrus at the console.

# Developer Notes
I originally wanted to just call `journalctl` from within the main program.
The problem I ran into was that `journalctl` seemed like it was
checking its own calling name and trying to match it with some known
program names.
It was erroring out with some "Failed to match name blahhhh".


[logrus]: https://github.com/sirupsen/logrus
[journalhook]: https://github.com/wercker/journalhook
[journald-fields]: https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html