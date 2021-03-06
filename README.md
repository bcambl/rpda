rpda
====
Is a RecoverPoint Direct Access utility to provide clients a simple CLI to manage _direct access_ testing on the latest image of a copy node within a consistency group remotely via API. The project is written in Go (golang) and can be compiled to a single binary for ease of deployment.

**DOWNLOAD THE LATEST VERSION OF THIS UTILITY ON THE RELEASE PAGE [HERE](https://github.com/bcambl/rpda/releases/latest)**

## User Permissions
An account on the RecoverPoint Appliance is required and the user must have access to administrate desired consistency groups. This utility will only administer consistency groups of which the account has access to modify as per RecoverPoint user privledges when `--all` is used.

## Configuration
A configuration template will be generated upon first execution of `rpda [command]`. (ie: `rpda list` or `rpda status --all`)

```
api:
  delay: 0
  polldelay: 3
  pollmax: 30
  url: https://recoverpoint_fqdn/
  username: username
identifiers:
  copy_node_regexp: _CN$
  production_node_regexp: _PN$
  test_node_regexp: ^TC_

```

Update the configuration file with variables that suit your site or environment.  
The `identifiers` section in the configuration file uses [_regexp_](https://golang.org/pkg/regexp/) to determine the desired copy for when `--test` or `--dr` are used.

The following example will work with the default `identifiers` section in the configuration example above.

```
EXAMPLE_CONSISTENCY_GROUP_CG:

             PRODUCTION_NODE_PN          <-- Production/Protection node
                /        \
               /          \
              /            \
             /              \
            /        TC_TESTING_COPY_CN  <-- copy node for testing or long term direct access
           /
       COPY_NODE_CN  <---------------------- un-interupted copy node for disaster recovery
```

## Available Commands
- `list`    List all Consistency Group Names
- `status`  Display Consistency Group Status
- `enable`  Enable direct access mode for the latest copy
- `finish`  Return a conistency group to a full replication state
- `help`    Help about any command

## Specifying a Copy
Naming consistency groups using a consistent _naming scheme_ will allow the use of `--test` and `--dr` options by configuring the `identifiers` section with regular expressions to suite your environment. _(see configuration section above)_

One of the following _copy flags_ must be provided:
 - `--copy <copy_name>` to specify copy name to enable direct access
 - `--test` to use the latest _test_ copy based on `test_node_regexp` regular expression within the configuration file
 - `--dr` to use the latest _test_ copy based on `copy_node_regexp` regular expression within the configuration file

 Note:  
 - only one of the above flags can be provided at once.
 - `--copy` cannot be combined with `--all`

## Additional Flags

- `--user <username>` will override the `username` specified within the configuration file. _(will prompt for password)_
- `--delay 60`: will introduce a delay of `60` seconds between consistency group changes when using `--all` (default: `0`)
- `--polldelay 10`: will modify the seconds which the utility will wait between API status polling requests (default: `3`)
- `--pollmax 60`: will modify the number of status poll attempts with before failing (default: `30`)
- `--debug`: will produce additional debugging output to assist with troubleshooting & development
- `--check`: will run allow the application to execute _without_ making any changes (`GET` requests only)
- `--help`: will display CLI help and examples

  Note:  
  Each _command_ also has its own `--help`  
  ```
  rpda enable --help
  ```

## Command-Line Examples

### List  
List All Consistency Groups
```
rpda list
```

### Status  
Display Status of all Consistency Groups
```
rpda status --all
```

Display Status of Consistency Group `TestGroup_CG`
```
rpda status --group TestGroup_CG
```

### Enable Direct Access  
Enable Direct Image Access Mode for the **_Test_ Copy** on **_ALL_** Consistency Groups
```
rpda enable --all --test
```

Enable Direct Image Access Mode for the **_Test_ Copy** on **_ALL_** Consistency Groups with `30` second **delay**
```
rpda enable --all --test --delay 30
```

Enable Direct Image Access Mode for the **_Test_ Copy** on Consistency Group `TestGroup_CG`
```
rpda enable --group TestGroup_CG --test
```

Enable Direct Image Access Mode for the **_DR_ Copy** on **_ALL_** Consistency Groups
```
rpda enable --all --dr
```

Enable Direct Image Access Mode for the **_DR_ Copy** on Consistency Group `TestGroup_CG`
```
rpda enable --group TestGroup_CG --dr
```

Enable Direct Image Access Mode for a **_User Defined Copy_** (`_Example_CN_`) on Consistency Group `TestGroup_CG`
```
rpda enable --group TestGroup_CG --copy Example_CN
```

### Finish Testing (Disable Direct Access & Start Tansfer)
Finish Direct Image Access Mode on **_ALL_** Consistency Groups for **_Test_ Copy**
```
rpda finish --all --test
```

Finish Direct Image Access Mode on **_ALL_** Consistency Groups for **_Test_ Copy** with `60` second delay
```
rpda finish --all --test --delay 60
```

Finish Direct Image Access Mode on Consistency Group `TestGroup_CG` for **_Test_ Copy**
```
rpda finish --group TestGroup_CG --test
```

Finish Direct Image Access Mode on **_ALL_** Consistency Groups for **_DR_ Copy**
```
rpda finish --all --dr
```

Finish Direct Image Access Mode on Consistency Group `TestGroup_CG` for **_DR_ Copy**
```
rpda finish --group TestGroup_CG --dr
```

Finish Direct Image Access Mode on Consistency Group `TestGroup_CG` for **_User Defined Copy_** (`_Example_CN_`)
```
rpda finish --group TestGroup_CG --copy Example_CN
```

## Build Instructions

**DOWNLOAD THE LATEST VERSION OF THIS UTILITY ON THE RELEASE PAGE [HERE](https://github.com/bcambl/rpda/releases/latest)**

If you would rather compile the project from source, please install the latest version of the Go programming language  [here](https://golang.org/dl/).

Build Project for `x86_64/amd64` Linux
```
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o rpda main.go
```

Build Project for `x86_64/amd64` Windows
```
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o rpda.exe main.go
```

Please read more about building the project for other platforms (`$GOOS` and `$GOARCH`) [here](https://golang.org/doc/install/source#environment).
