<?php
$pdo = new PDO("sqlite::memory:");
$pdo->exec("create table item (id integer primary key, name text)");
var_dump($pdo->inTransaction());

$pdo->beginTransaction();
var_dump($pdo->inTransaction());
$pdo->exec("insert into item (name) values ('rolled back')");
$pdo->rollBack();
var_dump($pdo->inTransaction());
var_dump($pdo->query("select count(*) from item")->fetchColumn());

$pdo->beginTransaction();
$pdo->exec("insert into item (name) values ('kept')");
$pdo->commit();
var_dump($pdo->inTransaction());
var_dump($pdo->query("select count(*) from item")->fetchColumn());

$pdo->exec("SAVEPOINT outer");
$pdo->exec("insert into item (name) values ('before savepoint')");
$pdo->exec("SAVEPOINT inner");
$pdo->exec("insert into item (name) values ('discarded')");
$pdo->exec("ROLLBACK TO SAVEPOINT inner");
$pdo->exec("RELEASE SAVEPOINT outer");
var_dump($pdo->query("select name from item order by id")->fetchAll(PDO::FETCH_COLUMN));

try {
    $pdo->rollBack();
} catch (PDOException $e) {
    echo get_class($e), "\n";
}
