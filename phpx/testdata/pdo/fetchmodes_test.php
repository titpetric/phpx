<?php
// The fetch modes php answers beyond a flat row: grouping, unique keys, a
// callback, a class and an object, plus the errors each raises.
$pdo = new PDO("sqlite::memory:");
$pdo->exec("create table t (id integer primary key, g text, name text)");
$pdo->exec("insert into t (g, name) values ('a', 'one'), ('a', 'two'), ('b', 'three')");
$q = fn() => $pdo->query("select g, name from t order by id");

print_r($q()->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC));
print_r($q()->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_NUM));
print_r($q()->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_OBJ));
print_r($q()->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_COLUMN));
print_r($q()->fetchAll(PDO::FETCH_UNIQUE | PDO::FETCH_ASSOC));

print_r($q()->fetchAll(PDO::FETCH_FUNC, fn($g, $name) => "$g=$name"));
try {
    $q()->fetchAll(PDO::FETCH_FUNC);
} catch (ArgumentCountError $e) {
    echo get_class($e), ": ", $e->getMessage(), "\n";
}

print_r($pdo->query("select id, name from t order by id")->fetchAll(PDO::FETCH_KEY_PAIR));
try {
    $q()->fetchAll(PDO::FETCH_KEY_PAIR);
} catch (PDOException $e) {
    echo get_class($e), ": ", $e->getMessage(), "\n";
}

print_r($pdo->query("select g as x, name as x from t limit 1")->fetchAll(PDO::FETCH_NAMED));
print_r($pdo->query("select g, name from t limit 1")->fetchAll(PDO::FETCH_NAMED));

class Row
{
    public $g;
    private $name;

    public function __construct()
    {
        $this->g = strtoupper($this->g);
    }

    public function name()
    {
        return $this->name;
    }
}

$row = $pdo->query("select g, name from t limit 1")->fetchAll(PDO::FETCH_CLASS, "Row")[0];
echo $row->g, " ", $row->name(), "\n";

$target = new stdClass();
$into = $pdo->query("select g, name from t limit 1");
$into->setFetchMode(PDO::FETCH_INTO, $target);
$into->fetch();
echo $target->g, " ", $target->name, "\n";

$byClass = $pdo->query("select g, name from t limit 1");
$byClass->setFetchMode(PDO::FETCH_CLASS, 'Row');
$row = $byClass->fetch();
echo get_class($row), " ", $row->g, " ", $row->name(), "\n";


$statement = $pdo->query("select g from t order by id");
$statement->setFetchMode(PDO::FETCH_COLUMN, 0);
echo $statement->fetch(), " ", $statement->fetch(), " ", $statement->fetch(), "\n";

echo "rowCount: ", var_export($q()->rowCount(), true), "\n";
