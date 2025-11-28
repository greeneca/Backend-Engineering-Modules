#! /bin/bash
set -e

cqlsh $1 -e "CREATE KEYSPACE IF NOT EXISTS wiki_updates WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"
