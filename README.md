rpda
====
is a RecoverPoint Direct Access utility to enable _direct access_ to the latest image on a copy node
within a consistency group remotely via API. 
The project is written in Go (golang) and can be compiled to a single binary for ease of deployment.

Download latest compiled `x86_64` release binary [here](https://github.com/bcambl/rpda/releases)

## Configuration
A configuration template will be generated upon first execution of `rpda`. 

```
api:
  delay: 0
  url: https://recoverpoint_fqdn/
  username: username
identifiers:
  copy_node_regexp: _CN$
  production_node_regexp: _PN$
  test_node_regexp: ^TC_


```

Update the configuration file with variables that suit your site or environment.  
Ensure the `identifiers` section within the configuration file is updated as these values are used
to identify the copies by name within the consistency groups.

```
EXAMPLE_CONSISTENCY_GROUP_CG:

             PRODUCTION_NODE_PN          <-- Production instance
                /        \
               /          \
              /            \
             /              \
            /        TC_TESTING_COPY_CN  <-- copy for testing or long term direct access
           /
       COPY_NODE_CN  <---------------------- un-interupted copy for disaster recovery 
```

Please reference the above example with the default `identifiers` section in the configuration.  
This utility relies upon the use of [_regexp_](https://golang.org/pkg/regexp/) to determine the desired copy.

## User Permissions
An account on the RecoverPoint Appliance is required and the user must have access to administrate desired consistency groups.
When issuing the `--all` option will only administer consistency groups of which the account has access to modify as per RecoverPoint user privledges.

## Specifying a Copy
As mentioned above, if the consistency groups being managed have a consistent naming structure, you can configure the
`identifiers` with appropriate regular expressions within the configuration file (default location: `$HOME/.rpda.yaml`).

One of the following _copy flags_ must be provided:
 - `--copy <copy_name>` to specify copy name
 - `--test` to use the latest _test_ copy based on `test_node_regexp` regular expression within the configuration file
 - `--dr` to use the latest _test_ copy based on `copy_node_regexp` regular expression within the configuration file

 Note:  
 - only one of the above flags can be provided at once.
 - `--copy` cannot be combined with `--all`

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

Enable Direct Image Access Mode for a **_User Defined Copy_** (_Example_CN_) on Consistency Group `TestGroup_CG`
```
rpda enable --group TestGroup_CG --copy Example_CN
```

### Finish Testing (Disable Direct Access & Start Tansfer)
Finish Direct Image Access Mode on **_ALL_** Consistency Groups for **_Test_ Copy**
```
rpda finish --all --test
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

Finish Direct Image Access Mode on Consistency Group `TestGroup_CG` for **_User Defined Copy_** (_Example_CN_)
```
rpda finish --group TestGroup_CG --copy Example_CN
```

## Additional Flags

- `--debug`: will print various important variables and run various debugging functions
- `--check`: will run allow the application to execute _without_ making any changes (readonly mode)

## Project Dependencies/Libraries
The following libraries were used in the creation of this project:  

-	`github.com/mitchellh/go-homedir`-> home directory discovery
-	`github.com/sirupsen/logrus`     -> enhanced logging output
-	`github.com/spf13/cobra`         -> standard cli framework
-	`github.com/spf13/viper`         -> configuration file management
-	`golang.org/x/crypto`            -> password prompt

## Build Instructions

Download latest compiled `x86_64/amd64` release binary [here](https://github.com/bcambl/rpda/releases)

If you would like to compile the project from source, please install the latest version of the
Go programming [here](https://golang.org/dl/).

Build Project for `x86_64/amd64` Linux
```
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o rpda main.go
```

Build Project for `x86_64/amd64` Windows
```
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o rpda.exe main.go
```

Please read more about building the project for other platforms ($GOOS and $GOARCH) [here](https://golang.org/doc/install/source#environment).
