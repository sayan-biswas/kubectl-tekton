# kubectl-tekton
kubectl-tekton is a plugin for kubectl CLI for fetching tekton resources from tekton results.

## Installation

Easiest way to install the plugin is using Krew plugin manager.
```shell
kubectl krew install --manifest-url https://raw.githubusercontent.com/sayan-biswas/kubectl-tekton/main/.krew.yaml
```

## Usage

### Configuration

Run the following command to configure tekton results instance
```shell
kubectl tekton config results
```

To reconfigure use the following flag
```shell
kubectl tekton config results -r
```

The configuration is interactive and the input is stored in kubeconfig context. This enables storing tekton results configuration per context.

In Openshift platform with proper access the CLI can discover installed tekton results instances.

**NOTE:**
Use gRPC client as on now, the REST client is incomplete, as filters serialization is not yet implemented.

### Fetching Resources

To list PipelineRuns in the namespace
```shell
kubectl tekton get pr -n default
```

To list PipelineRuns with a specific name in the namespace
```shell
kubectl tekton get pr testpr -n default
```

To print a PipelineRun in the namespace
```shell
kubectl tekton get pr testpr -n default -o yaml
```
```
--uid       flag can be used to specify a particular resource
--output    can be used to print the resource in JSON and YAML
--limit     can be used to the number of items
```

**NOTE:**
If UID flag is not specified the last updated resource will be printed

### Fetching Logs

To list PipelineRuns in the namespace
```shell
kubectl tekton log pr testpr -n default
```