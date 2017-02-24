# clusterping
The tool to ping cluster nodes (Suitable for Percona XtraDB Cluster, Group Replication etc) and measure latency of update operation

How to use:

`./clusterping -host=10.10.7.165:3306 -nodes=10.10.7.164:3306,10.10.7.167:3306 -user=root -password=Theistareyk `

where `host` - the primary node, where we perform update operation, `nodes` - list of nodes to measure delay in propagation
