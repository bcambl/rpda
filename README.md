rpda
====
is a RecoverPoint Direct Access utility to enable _direct access_ to the latest image on a copy node
within a consistency group remotely via API. 
The project is written in Go (golang) and can be compiled to a single binary for ease of deployment.

Download latest compiled `x86_64` release binary [here](#)

## Configuration
A configuration template will be generated upon first execution of `rpda`. 

```
api:
  delay: 0
  url: https://recoverpoint_fqdn/
  username: username
identifiers:
  dr_copy_name_contains: _CN
  production_node_name_contains: _PN
  test_copy_name_contains: TC_

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
This utility relies upon the use of a (_strings.Contains_)[https://golang.org/pkg/strings/#Contains]
to determine the desired copy.

## User Permissions
An account on the RecoverPoint Appliance is required and the user must have access to administrate all desired consistency groups.  
Note that when issuing the `--all` option will only administer consistency groups of which the account has access to modify as per RecoverPoint user privledges.

## Command-Line Examples

List All Consistency Groups
```
rpda list
```

Display Status of all Consistency Groups
```
rpda status --all
```

Display Status of Consistency Group `TestGroup`
```
rpda status --group TestGroup
```

Enable Direct Image Access Mode for the **_Test_ Copy** on _ALL_ Consistency Groups
```
rpda start --all --test
```

Enable Direct Image Access Mode for the **_Test_ Copy** on Consistency Group `TestGroup` 
```
rpda start --group TestGroup --test
```

Enable Direct Image Access Mode for the **_DR_ Copy** on _ALL_ Consistency Groups
```
rpda start --all --dr
```

Enable Direct Image Access Mode for the **_DR_ Copy** on Consistency Group `TestGroup` 
```
rpda start --group TestGroup --dr
```

Finish Direct Image Access Mode on _ALL_ Consistency Groups
```
rpda finish --all
```

Finish Direct Image Access Mode on Consistency Group `TestGroup`
```
rpda finish --group TestGroup
```

## Build Instructions

Download latest compiled `x86_64` release binary [here](#)

If you would like to compile the project from source, please install the latest version of the
Go programming [here](https://golang.org/dl/).

Build Project for `x86_64` Linux
```
GOOS=linux go build -ldflags="-s -w" -o rpda main.go
```

Build Project for x86_64 Windows
```
GOOS=windows go build -ldflags="-s -w" -o rpda main.go
```

Please read more about building the project for other platforms [here](https://golang.org/pkg/go/build/).