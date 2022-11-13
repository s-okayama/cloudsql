# cloudsql
## Overview
Save time for gcloud sql instances list and cloud_sql_proxy command.

## Install
```
- Preparation　
1. Install Cloud SQL Auth Proxy 
https://cloud.google.com/sql/docs/postgres/sql-proxy?hl=ja#install
$ chmod +x cloud_sql_proxy
$ sudo mv cloud_sql_proxy /usr/local/bin/

2. Install Gcloud CLI
https://cloud.google.com/sdk/docs/install?hl=ja

3. Check Preparation
$ gcloud --version
$ cloud_sql_proxy --version

- Insall
Chose Download or Build
# Download
https://github.com/s-okayama/cloudsql/releases

# Build
$ git clone git@github.com:s-okayama/cloudsql.git
$ go build
$ chmod +x cloudsql
$ sudo mv cloudsql /usr/local/bin/

- Set Config File
$ mkdir ~/.cloudsql/
$ vim ~/.cloudsql/config
project-dev
project-prd 
```

## Usage
```
##### help #####
$ cloudsql        
CloudSQL CLI

Usage:
  cloudsql [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  connect     connect to cloudsql instance
  disconnect  disconnect cloudsql instance
  help        Help about any command

Flags:
  -h, --help   help for cloudsql


##### connect #####
$ cloudsql connect
Use the arrow keys to navigate: ↓ ↑ → ← 
? Select Project: 
    project-prd
  ▸ project-dev
✔ project-dev

? Select Project:project-dev 
  ▸ project-dev:asia-northeast1:stg-hoge-db-fecdf019
    project-dev:asia-northeast1:stg-postgres-0e80e42e
    project-dev:asia-northeast1:stg-mysql-db-8347a466
    project-dev:asia-northeast1:stg-metabase-db-3413a639
 ✔ project-dev:asia-northeast1:stg-postgres-db-0e80e42e

? Select Database:
  ▸ postgres
    test-db 
Connecting Instance
2022/11/06 21:21:45 Cloudsql proxy process is running in background, process_id: 65464
Can connect using:
psql -h localhost -U yamada.taro@gmail.com -p 5432 -d postgres


##### disconnect #####
$ cloudsql disconnect          
Use the arrow keys to navigate: ↓ ↑ → ← 
? Select Instance to disconnect: 
  ▸ project-dev:asia-northeast1:stg-postgres-0e80e42e=tcp:5432
```

## ToDo
- [x] Disable sound for Mac  
→ Add nobell.go
- [x] Add search feature
→ Can you search by /
- [x] Add Select Database feature  
→ Add getDatabase & get listDatabase func
- [ ] Add proxy & connect mode
- [ ] Add Doctor feature(check cloud_sql_proxy & postgres & mysql)
