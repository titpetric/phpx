<?php
$pdo = new PDO("sqlite::memory:");
var_dump(get_class($pdo), $pdo instanceof PDO, PDO::ATTR_ERRMODE, PDO::FETCH_OBJ);
$pdo->exec("create table item (id integer primary key, name text, price real, note text)");
echo $pdo->exec("insert into item (name, price, note) values ('first', 2.5, NULL)"), "\n";

$insert = $pdo->prepare("insert into item (name, price) values (:name, :price)");
$insert->bindValue(":name", "second", PDO::PARAM_STR);
$insert->bindValue(":price", 4, PDO::PARAM_INT);
var_dump($insert->execute());
echo $pdo->lastInsertId(), "\n";

$select = $pdo->prepare("select id, name, price, note from item where price > ? order by id");
$select->bindValue(1, 3, PDO::PARAM_INT);
var_dump($select->execute(), $select->columnCount(), $select->rowCount());
print_r($select->fetch(PDO::FETCH_ASSOC));
print_r($select->fetch(PDO::FETCH_NUM));
print_r($select->fetchAll(PDO::FETCH_OBJ));
print_r($select->fetchAll());

$query = $pdo->query("select id, name from item order by id");
print_r($query->fetchAll(PDO::FETCH_KEY_PAIR));
$query = $pdo->query("select name from item order by id");
print_r($query->fetchAll(PDO::FETCH_COLUMN));

var_dump($pdo->quote("O'Reilly"));
print_r($pdo->getAttribute(PDO::ATTR_DRIVER_NAME));
var_dump($pdo->errorCode(), $pdo->errorInfo());
