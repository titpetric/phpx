<?php
class Bag {
    public int $n {
        set {
            echo "set n\n";
            $this->n = $value * 2;
        }
    }
    public string $name = "x" {
        get => strtoupper($this->name);
    }
    public string $full {
        get => $this->first . " " . $this->last;
        set(string $v) {
            [$this->first, $this->last] = explode(" ", $v);
        }
    }
    public private(set) int $count = 0;
    private string $first = "";
    private string $last = "";
    public function bump(): void { $this->count++; }
}
$b = new Bag;
$b->n = 4;
echo $b->n, "\n";
echo $b->name, "\n";
$b->full = "Ada Lovelace";
echo $b->full, "\n";
$b->bump();
echo $b->count, "\n";
