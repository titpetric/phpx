<?php
$path = tempnam(sys_get_temp_dir(), 'phpx_pdo_');
$pdo = new PDO('sqlite:' . $path);
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec('CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)');
var_dump($pdo->inTransaction());
$pdo->exec('BEGIN');
var_dump($pdo->inTransaction());
$pdo->exec("INSERT INTO t (name) VALUES ('a')");
$pdo->exec('ROLLBACK');
var_dump($pdo->inTransaction());
$count = $pdo->query('SELECT COUNT(*) FROM t')->fetchColumn();
var_dump($count);
unlink($path);
