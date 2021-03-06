

![Version: 0.27.0](https://img.shields.io/badge/Version-0.27.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| K8ssandra Team | k8ssandra-developers@googlegroups.com | https://github.com/k8ssandra |

## Source Code

* <https://github.com/k8ssandra/k8ssandra>
* <https://github.com/k8ssandra/k8ssandra/tree/main/charts/restore>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| name | string | `"restore"` | Name of the CassandraRestore custom resource |
| backup.name | string | `"backup"` | Name of the CassandraBackup custom resource to be restored from |
| cassandraDatacenter.name | string | `"dc1"` | Name of the CassandraDatacenter where the CassandraBackup will be restored |
| inPlace | bool | `true` | In-place restore will restore the backup to the source cluster. Note that this will trigger a rolling restart of the cluster. |
| shutdown | bool | `false` | When true will shutdown the entire Cassandra cluster. The underlying StatefulSets are scaled down to zero. Persistent volumes remain intact. If the backup includes schema changes like dropping a table, then set this to true; otherwise, the changes will be lost via gossip from nodes that have not yet been restored. |
