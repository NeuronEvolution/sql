### start
docker run --name mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=123456 -v /Users/god/work/mysql-data:/var/lib/mysql -d

### Attention
    - mysql connection string with ?parseTime=true
    
### Feather
    - query builder
    - golang sql wrapper
    - transaction
    - extendable
    - upgrade ddl shell
    
### Todo
    - metric
    - batch
    - simple join,may be no need

### #########################    
    
### 增删改
    - update width version//根据主键全部更新，带版本号
        -- UPDATE table set all,update_version=$update_version+1 where id=? AND update_version=$update_version
    - partial update//根据主键部分更新
        -- UPDATE table set c1=?,c2=? WHERE id= ?
    - partial update with version////根据主键部分更新，带版本号
        -- UPDATE table set c1=?,c2=?,update_version=$update_version+1 WHERE id= ? AND update_version=$update_version

### 其它
    - 参数化，防sql注入
    - 字符串截断检测
    - 自动生成Statement
    
# INSERT
INSERT 
    [INTO] tbl_name
    [(col_name [, col_name] ...)]
    {VALUES | VALUE} (value_list) [, (value_list)] ...
    [ON DUPLICATE KEY UPDATE assignment_list]
    
    INSERT [ON DUPLICATE]
    BATCH INSERT [ON DUPLICATE]
# DELETE
DELETE FROM tbl_name [WHERE where_condition]
    
    DELETE by id
    DELETE by where 
# UPDATE
UPDATE table_reference
    SET assignment_list
    [WHERE where_condition]
    
    UPDATE all by id
    UPDATE patially by id
    UPDATE patially by where
# SELECT
SELECT
    [DISTINCT]
    select_expr [, select_expr ...]
    FROM table_references
    [WHERE where_condition]
    [GROUP BY col_name[ASC | DESC], ...]
    [ORDER BY col_name[ASC | DESC], ...]
    [LIMIT {[offset,] row_count | row_count OFFSET offset}]
    [FOR {UPDATE | SHARE}
    
    SELECT by id
    SELECT * where
    
    SELECT JOIN
    
# PARTITION