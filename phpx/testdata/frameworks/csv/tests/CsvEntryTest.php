<?php
use PHPUnit\Framework\TestCase;
use League\Csv\Reader;
use League\Csv\Writer;

final class CsvEntryTest extends TestCase
{
    public function testReadWithHeader(): void
    {
        $csv = Reader::createFromString("name,age\nann,30\nbob,25\n");
        $csv->setHeaderOffset(0);
        $rows = iterator_to_array($csv->getRecords(), false);
        $this->assertSame([['name' => 'ann', 'age' => '30'], ['name' => 'bob', 'age' => '25']], $rows);
        $this->assertSame(['name', 'age'], $csv->getHeader());
    }
    public function testWrite(): void
    {
        $w = Writer::createFromString();
        $w->insertOne(['a', 'b c', 'd"e']);
        $w->insertAll([[1, 2, 3], ['x', 'y,z', '']]);
        $this->assertSame("a,\"b c\",\"d\"\"e\"\n1,2,3\nx,\"y,z\",\n", str_replace("\r\n", "\n", $w->toString()));
    }
    public function testCount(): void
    {
        $csv = Reader::createFromString("a\nb\nc\n");
        $this->assertSame(3, count($csv));
    }
}
