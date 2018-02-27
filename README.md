### start
docker run --name mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=123456 -v /Users/god/work/mysql-data:/var/lib/mysql -d

### Todo
    - metric
    - batch
    - simple join,may be no need
    - upgrade ddl shell

### Attention
    - mysql connection string with ?parseTime=true
### Feather
    - query builder
    - golang sql wrapper
    - lo
    - transaction
    - extendable
    - where
    - count
    - group by
    - order by