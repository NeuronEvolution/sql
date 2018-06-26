### start
docker run --name mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=123456 -v /Users/god/work/mysql-data:/var/lib/mysql -d

### Attention
    - mysql connection string with ?parseTime=true
    
### Todo
    - metric
    - join
    - 字符串截断检测
    - 自动生成Statement
    - onduplicated key update 指定更新字段
    - 增加对update_time的自动输入
    - ［已完成］优化limit
    
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