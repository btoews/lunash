# lunash

These are tools for configuring a Safenet Luna HSM. Configuration is normally done over SSH using a non-standard shell. This makes it difficult to automate configuration. For example, you cannot send two SSH commands like `ssh admin@myhsm -- hsm login -p s3cret; partition create ...`.  The tools in this repo address this problem by implementing SSH/SCP clients, allowing you to automate the configuring of your HSMs.

## Configuration

These tools are told about HSM hostnames, ports, and passwords from a config file. See [example_lunash.json](example_lunash.json) for an example of a configuration file. You are allowed to specify a "nickname" for each HSM in the config. Commands that accept a `-name` or `-names` parameter will accept either the hostname or the nickname of a given HSM.

## Tools

### `lunash`

The `lunash` command runs a series of SSH commands on one or more HSMs, optionally running `hsm login -p ${password}` first.

#### Examples:

Run `user list` followed by `partition list` on all HSMs in the config file:
```bash
bin/lunash -all -command "user list; partition list"
```


Login (`hsm login`) with the password in the config file and then run `partition create` on `hsm1.mycorp.net`:
```bash
bin/lunash -login=true -names hsm1.mycorp.net -command "partition create -partition pname -label plabel -password ppassword -domain pdomain -f"
```

### `lunascp-get`

The `lunascp-get` command SCP's a file from the HSM, outputting it to stdout.

#### Examples:

Get `server.pem` from the HSM with nickname `hsm1`:

```bash
bin/lunascp-get -name hsm1 -path server.pem > server.pem
```

### `lunascp-put`

The `lunascp-put` command SCP's a file from stdin to the HSM.

#### Examples:

Put `client.pem` on the HSM with nickname `hsm1`:

```bash
bin/lunascp-put -name hsm1 -path client.pem < client.pem
```
