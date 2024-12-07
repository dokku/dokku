# Procfile

A Procfile is a file that was [promoted by Heroku](https://blog.heroku.com/the_new_heroku_1_process_model_procfile) for their platform as an easy way to specify one or more distinct processes to run within Heroku. This format has since been picked up by various other tools and platforms.

## General Overview

The `procfile-util` tool expects a Procfile to be defined as one or more lines containing one of:

- a comment (preceeded by a `#` symbol or two `//` characters)
- a process-type/command combination (with optional trailing whitespace or trailing comment)
    - when there is a trailing comment, the `#` symbol/`//` characters _must_ be preceeded by one or more `whitespace` characters.
- a blank line (with optional trailing whitespace)

Comments and blank lines are ignored, while process-type/command combinations look like the following:

```
<process type>: <command>
```

The syntax is defined as follows:

- `<process type>` – any character in the class `[A-Za-z0-9_-]+`, a process type is a name for your command, such as `web`, `worker`, `urgentworker`, `clock`, etc.
- `<command>` – a command used to launch the process, such as `rake jobs:work`

Additional rules are as follows:

- contain at most 63 characters

A Procfile that does not match any of the above rules will be considered invalid, and not be processed. This is to avoid issues where a Procfile may contain merge conflicts or other improper content, thus resulting in unwanted runtime behavior for applications.

Finally, process types within a Procfile may not overlap and must be unique. Rather than assuming that the last or first specified is correct, `procfile-util` will fail to parse the Procfile with the relevant error.

### Strict Mode

> Strict mode can be triggered on `procfile-util` via the `--strict` flag.

In strict mode, the character set of a process type changes.

- `<process type>` – a valid DNS Label Name as per [RFC 1123](https://tools.ietf.org/html/rfc1123), a process type is a name for your command, such as `web`, `worker`, `urgentworker`, `clock`, etc.

This syntax differs common interpretations of valid `<process type>` values in that we define the process type as a DNS Label name, versus the regex `[A-Za-z0-9_]+`. The reason for this is that processes defined within Procfiles are commonly used in DNS entries. Rather than have a second level of platform-specific validation in place, this project implicitly defines the format for the process-type.

Given the above, a valid process type can be generalized to the following rules (as taken from the [Kubernetes documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names)):

- contain at most 63 characters
- contain only lowercase alphanumeric characters or '-'
- start with an alphanumeric character
- end with an alphanumeric character
