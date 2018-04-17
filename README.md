### start
docker run --name mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=123456 -v /Users/god/work/mysql-data:/var/lib/mysql -d

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
    - partial update
    - upgrade ddl shell
    
### Todo
    - metric
    - batch
    - simple join,may be no need