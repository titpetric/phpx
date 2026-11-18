<?php
$pdo = new PDO("sqlite::memory:");
var_dump($pdo->getAttribute(PDO::ATTR_ERRMODE));
$pdo->exec("create table item (id integer primary key, name text not null)");
$pdo->exec("insert into item (id, name) values (1, 'first')");

try {
    $pdo->exec("insert into item (id, name) values (1, 'again')");
} catch (PDOException $e) {
    echo get_class($e), "\n";
    echo $e->getMessage(), "\n";
    var_dump($e->getCode());
    var_dump($e->errorInfo);
}
try {
    $pdo->exec("select * from missing");
} catch (PDOException $e) {
    echo $e->getMessage(), "\n";
    var_dump($e->errorInfo);
}
try {
    $statement = $pdo->prepare("select * from item where id = ?");
    $statement->execute();
} catch (PDOException $e) {
    echo $e->getMessage(), "\n";
}

$silent = new PDO("sqlite::memory:");
$silent->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_SILENT);
var_dump($silent->query("select * from missing"));
var_dump($silent->errorCode());
var_dump($silent->errorInfo());

