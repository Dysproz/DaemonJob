# DaemonJob

## Idea
This project is the response to problem that from time to time every Kubernetes developer come across - Resource that works like DaemonSet, but triggers Jobs instead.

Kubernetes from years hasn't resolved multiple issues submitions (like [this](https://github.com/kubernetes/kubernetes/issues/64623) one) so working on custom resource for [contrail-operator](https://github.com/Juniper/contrail-operator) I've came with an idea of kubernetes operator that implements custom resource DaemonJob.

## Implementation
This is Kubernetes operator written in Go with support of operator-sdk so it managing pod running in target cluster as well as apply manifests with CRDs, RBAC etc.

Logic of DaemonJob is that all parameters are declared as for standard Job resource except for *parrarel* and *completions* fields which are filled based on number of applicable nodes (which you may control with for example nodeSelector).
DaemonJob has Anti-affinity added so that on single node only one pod will be run.
Having that connected together we achieve pretty much logic of DaemonSet.

The only disadvantage is restrictive policy of Job resource which does not allow to edit *completions* or *parrarel* fields on the go (or even a lot of pod spec values). Because of that with every such change DaemonJob has to delete and create new Job.

## Deployment
This repository contains useful *Makefile*.
In order to apply all required manifests onto cluster just run `make install`.

Afterwards, you'll need Pod with a manager running.
You may create one with:
```
export IMG=dysproz/daemon-job
make deploy
```
**NOTE**: You may of course apply your own image (for example with edits necessary for your project). In that case just export IMG as your image.

And that's it. Now you may create your own manifests for DaemonJob and apply them to the cluster.
Example manifest may be found under *config/samples/dj_v1_daemonjob.yaml*.

## Builing your own image
In case you want to edit this code, build and then push your own image you may do that by:
* exporting IMG variable `export IMG=[CONTAINER REPOSITORY]/[CONTAINER NAME]:[TAG]
* building operator with `make docker-build`
* pushing it to the registry (remember to first log in) `make docker-push`
