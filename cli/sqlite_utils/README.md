# SQLite utils

## Database exporter

The data exporter consumes a process ID then dry runs a lua script to read all the
database, or a target database, from the AOS-SQLITE process.

This is then written to a .sqlite file which can be used in any other sqlite database system.

Internally, this is base64 encoded for egress purposes.

## Database importer

The SQLite importer requires a wallet to sign, with Eval authority on the process (normally, the process owner). This allows you to add prebuilt databases to the process and when combined with the exporter, allows users to completely fork SQL state of processes.