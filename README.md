# kubectl-tekton
kubectl-tekton is a plugin for kubectl CLI for fetching tekton resources from tekton results.

## Installation

Easiest way to install the plugin is using Krew plugin manager.
```shell
kubectl krew index add test https://github.com/sayan-biswas/kubectl-tekton.git
```
```shell
kubectl krew install test/tekton
```

## Usage

### Configuration

#### List Of Configurations

| Name                     | Group  | Default | Description                            |
|--------------------------|--------|---------|----------------------------------------|
| host                     | client |         | Host address for the client to connect |
| client-type              | client | REST    | Type of client can be GRPC or REST     |
| insecure-skip-tls-verify | client | false   | Skip host name verification            |
| timeout                  | client | 1m      | Client context timeout                 |
| certificate-authority    | tls    |         | CA file path to use                    |
| client-certificate       | tls    |         | Certificate file path to use           |
| client-key               | tls    |         | Key file path to use                   |
| tls-server-name          | tls    |         | Override hostname for TLS              |
| token                    | auth   |         | Token to use for authorization         |
| act-as                   | auth   |         | User ID for impersonation              |
| act-as-uid               | auth   |         | UID for impersonation                  |
| act-as-groups            | auth   |         | Groups for impersonation               |


Configure tekton results client
```shell
kubectl tekton config results
```

Reconfigure tekton results client
```shell
kubectl tekton config results --reset
```

To remove a value from config file
```shell
kubectl tekton config results token="" --prompt=false
```

Configure specific properties
```shell
kubectl tekton config results host token
```

Configure can also be called with the `group` name. Only configurations from that `group` will be prompted or configured.
```shell
kubectl tekton config results client
```

Non-interactive configuration (no validation).
```shell
kubectl tekton config results host="https://localhost:8080" token="test-token" --prompt="false"
```

Non-interactive prompt can also use the automated suggestions if the values are not provided
```shell
kubectl tekton config results host client-type --prompt="false"
```
This is update `host` and `client-type` from default or automated values.

The configuration is stored in kubeconfig context extension. This enables storing multiple tekton results configuration per context.

In Openshift platform with proper access the CLI can discover installed tekton results instances.

### Fetching Resources

List resources from a namespace
```shell
kubectl tekton get pr -n default
````

List limited resources from a namespace. By default only 10 resources are listed.
```shell
kubectl tekton get pr -n default --limit 20
```

Get resources by specifying name. Partial name can also be provided.
```shell
kubectl tekton get pr test -n default
```

Get resources by specifying UID. Partial UID can also be provided.
```shell
kubectl tekton get pr test -n default --uid="e0e4148c-b914"
```

List resources from a namespace with selectors. All the selectors support partial value.
```shell
kubectl tekton get pr -n default 
    --labels="app.kubernetes.io/name=test-app, app.kubernetes.io/component=database"
```

All selectors can be used together and works as AND operator.
```shell
kubectl tekton get pr -n default 
    --labels="app.kubernetes.io/name=test-app"
    --annotations="app.io/timeout=100"
```

All selectors except OwnerReferences can work with only key or value.
```shell
kubectl tekton get pr -n default --annotations="test" --labels="test"
```

Check if a particular annotation exists, without knowing the value.
```shell
kubectl tekton get pr -n default --annotations="results.tekton.dev/log"
```

OwnerReferences filter can not filter by key/value pair, but the filter should still be provided as key/value.
```shell
kubectl tekton get pr -n default --owner-references="kind=Service name=test-service"
```

Multiple owner references can be provided, but keys of each owner references should be seperated by space.
```shell
kubectl tekton get pr -n default 
    --owner-references="kind=Service name=test-service, kind=Deployment name=test-app"
```

OwnerReferences filter can be used to find child resources.
```shell
kubectl tekton get pr -n default --owner-references="name=parent-name"
```

Filter flag can be used to pass raw filter. Invalid syntax will cause error.
```shell
kubectl tekton get pr -n default --filter="data.status.conditions[0].reason in ['Failed']"
```

All the filters with partial string also. 

Flags
```
--uid               filter resource by UID
--output            print resource in JSON and YAML
--limit             limit number of items in page
--labels            filter resources by lables
--annotations       filter resources by annotations
--finalizers        filter resources by finalizers
--owner-references  filter resources by owner references
--filter            filter resources using raw filter string
```

### Fetching Logs

Get PipelineRun logs
```shell
kubectl tekton logs pr testpr -n default
```
Get TaskRun logs
```shell
kubectl tekton logs tr testtr -n default --uid="436dd41a-fd8a-4a29-b4f3-389b221af5dc"
```

### Deleting Resources

Delete all resources with name matching the keyword. All other flags for get are also available for delete.
Delete command will also delete any child resources matching the `OwnerReferences`.
```shell
kubectl tekton delete pr z56b6 -n default
```
