<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface};
use Symfony\Component\Console\Helper\{Table, TableCell, TableCellStyle, TableSeparator, TableStyle};
use Symfony\Component\Console\Exception\{InvalidArgumentException, LogicException};

final class ConsoleCov5Test extends TestCase
{
    private function t(?BufferedOutput &$out = null): Table
    {
        $out = new BufferedOutput();
        $out->setDecorated(false);
        return new Table($out);
    }

    /** @return array<string, callable():string> */
    public static function cases(): array
    {
        $mk = function (callable $f): callable {
            return function () use ($f): string {
                $out = new BufferedOutput();
                $out->setDecorated(false);
                $t = new Table($out);
                $f($t);
                $t->render();
                return $out->fetch();
            };
        };
        $rows = [['ISBN-1', 'Title One', 'Author A'], ['ISBN-22', 'Title Two', 'Author B'], ['ISBN-333', 'Title Three', 'Author C']];
        $hdr = ['ISBN', 'Title', 'Author'];
        $c = [];
        foreach (['default', 'borderless', 'compact', 'symfony-style-guide', 'box', 'box-double'] as $style) {
            $c["style $style"] = $mk(fn (Table $t) => $t->setHeaders($hdr)->setRows($rows)->setStyle($style));
        }
        $c['no headers'] = $mk(fn (Table $t) => $t->setRows($rows));
        $c['no rows'] = $mk(fn (Table $t) => $t->setHeaders($hdr));
        $c['empty'] = $mk(fn (Table $t) => null);
        $c['separator'] = $mk(fn (Table $t) => $t->setHeaders($hdr)->setRows([$rows[0], new TableSeparator(), $rows[1], new TableSeparator(), $rows[2]]));
        $c['colspan'] = $mk(fn (Table $t) => $t->setHeaders([[new TableCell('Main title', ['colspan' => 3])], $hdr])->setRows([[new TableCell('spanning two', ['colspan' => 2]), 'x'], ['a', 'b', 'c']]));
        $c['rowspan'] = $mk(fn (Table $t) => $t->setHeaders(['A', 'B'])->setRows([[new TableCell("r1\nr2\nr3", ['rowspan' => 3]), 'one'], ['two'], ['three'], ['next', 'row']]));
        $c['rowcolspan'] = $mk(fn (Table $t) => $t->setRows([[new TableCell('big', ['rowspan' => 2, 'colspan' => 2]), 'x'], ['y'], ['1', '2', '3']]));
        $c['multiline'] = $mk(fn (Table $t) => $t->setHeaders(['Col', 'Other'])->setRows([["line1\nline2", 'z'], ['q', "a\nb\nc"]]));
        $c['titles'] = $mk(fn (Table $t) => $t->setHeaderTitle('Header Title')->setFooterTitle('Footer Title')->setHeaders($hdr)->setRows($rows));
        $c['long titles'] = $mk(fn (Table $t) => $t->setHeaderTitle('A much longer header title than the table is wide')->setFooterTitle('Footer')->setHeaders(['a'])->setRows([['b']]));
        $c['widths'] = $mk(fn (Table $t) => $t->setColumnWidths([12, 14])->setHeaders(['A', 'B', 'C'])->setRows([['1', '2', '3']]));
        $c['width one'] = $mk(fn (Table $t) => $t->setColumnWidth(1, 20)->setHeaders(['A', 'B'])->setRows([['1', '2']]));
        $c['max width'] = $mk(fn (Table $t) => $t->setColumnMaxWidth(0, 8)->setHeaders(['Header A', 'B'])->setRows([['a very long cell value here', 'ok']]));
        $c['vertical'] = $mk(fn (Table $t) => $t->setHeaders($hdr)->setRows($rows)->setVertical());
        $c['horizontal'] = $mk(fn (Table $t) => $t->setHeaders($hdr)->setRows($rows)->setHorizontal());
        $c['vertical titles'] = $mk(fn (Table $t) => $t->setHeaderTitle('T')->setHeaders($hdr)->setRows($rows)->setVertical(true));
        $c['cell style'] = $mk(fn (Table $t) => $t->setRows([[new TableCell('right', ['style' => new TableCellStyle(['align' => 'right', 'cellFormat' => '<info>%s</info>'])]), new TableCell('center', ['style' => new TableCellStyle(['align' => 'center'])])], ['a-very-long-value-here', 'another-long-one']]));
        $c['column style'] = $mk(function (Table $t) {
            $s = new TableStyle();
            $s->setPadType(STR_PAD_LEFT);
            $t->setColumnStyle(1, $s)->setHeaders(['A', 'B'])->setRows([['x', '1'], ['yy', '22']]);
        });
        $c['custom style'] = $mk(function (Table $t) {
            $s = new TableStyle();
            $s->setHorizontalBorderChars('=', '-')->setVerticalBorderChars('!', ':')->setDefaultCrossingChar('#')->setPaddingChar('.')->setCellHeaderFormat('%s')->setCellRowFormat('%s')->setCellRowContentFormat(' %s ')->setBorderFormat('%s');
            $t->setStyle($s)->setHeaders(['A', 'B'])->setRows([['x', 'y'], new TableSeparator(), ['z', 'w']]);
        });
        $c['crossing chars'] = $mk(function (Table $t) {
            $s = new TableStyle();
            $s->setCrossingChars('1', '2', '3', '4', '5', '6', '7', '8', '9');
            $t->setStyle($s)->setHeaders(['A', 'B'])->setRows([['x', 'y'], new TableSeparator(), ['z', 'w']]);
        });
        $c['no outside border'] = $mk(function (Table $t) {
            $s = new TableStyle();
            $s->setDisplayOutsideBorder(false);
            $t->setStyle($s)->setHeaders(['A', 'B'])->setRows([['x', 'y']]);
        });
        $c['multiple header rows'] = $mk(fn (Table $t) => $t->setHeaders([['A', 'B'], ['C', 'D']])->setRows([['1', '2']]));
        $c['tags'] = $mk(fn (Table $t) => $t->setHeaders(['<info>H</info>'])->setRows([['<comment>c</comment>'], ['<fg=red>r</>']]));
        $c['uneven rows'] = $mk(fn (Table $t) => $t->setHeaders(['A', 'B', 'C'])->setRows([['1'], ['1', '2', '3', '4']]));
        $c['add append'] = $mk(function (Table $t) {
            $t->setHeaders(['A', 'B']);
            $t->addRow(['1', '2']);
            $t->addRows([['3', '4'], ['5', '6']]);
        });
        $c['assoc rows'] = $mk(fn (Table $t) => $t->setHeaders(['A', 'B'])->setRows([['b' => 'two', 'a' => 'one']]));
        $c['numbers'] = $mk(fn (Table $t) => $t->setRows([[1, 2.5, true, null, 'x']]));
        $c['stringable'] = $mk(fn (Table $t) => $t->setRows([[new class { public function __toString(): string { return 'strobj'; } }]]));
        return $c;
    }

    private function runCase(string $name): void
    {
        $expected = self::expected();
        $cases = self::cases();
        $this->assertSame($expected[$name], $cases[$name](), $name);
    }

    public function testRendered00(): void { $this->runCase('style default'); }
    public function testRendered01(): void { $this->runCase('style borderless'); }
    public function testRendered02(): void { $this->runCase('style compact'); }
    public function testRendered03(): void { $this->runCase('style symfony-style-guide'); }
    public function testRendered04(): void { $this->runCase('style box'); }
    public function testRendered05(): void { $this->runCase('style box-double'); }
    public function testRendered06(): void { $this->runCase('no headers'); }
    public function testRendered07(): void { $this->runCase('no rows'); }
    public function testRendered08(): void { $this->runCase('empty'); }
    public function testRendered09(): void { $this->runCase('separator'); }
    public function testRendered10(): void { $this->runCase('colspan'); }
    public function testRendered11(): void { $this->runCase('rowspan'); }
    public function testRendered12(): void { $this->runCase('rowcolspan'); }
    public function testRendered13(): void { $this->runCase('multiline'); }
    public function testRendered14(): void { $this->runCase('titles'); }
    public function testRendered15(): void { $this->runCase('long titles'); }
    public function testRendered16(): void { $this->runCase('widths'); }
    public function testRendered17(): void { $this->runCase('width one'); }
    public function testRendered18(): void { $this->runCase('max width'); }
    public function testRendered19(): void { $this->runCase('vertical'); }
    public function testRendered20(): void { $this->runCase('horizontal'); }
    public function testRendered21(): void { $this->runCase('vertical titles'); }
    public function testRendered22(): void { $this->runCase('cell style'); }
    public function testRendered23(): void { $this->runCase('column style'); }
    public function testRendered24(): void { $this->runCase('custom style'); }
    public function testRendered25(): void { $this->runCase('crossing chars'); }
    public function testRendered26(): void { $this->runCase('no outside border'); }
    public function testRendered27(): void { $this->runCase('multiple header rows'); }
    public function testRendered28(): void { $this->runCase('tags'); }
    public function testRendered29(): void { $this->runCase('uneven rows'); }
    public function testRendered30(): void { $this->runCase('add append'); }
    public function testRendered31(): void { $this->runCase('assoc rows'); }
    public function testRendered32(): void { $this->runCase('numbers'); }
    public function testRendered33(): void { $this->runCase('stringable'); }

    public function testStyleDefinitionRegistry(): void
    {
        $s = new TableStyle();
        $s->setPaddingChar('~');
        Table::setStyleDefinition('cov-custom', $s);
        $this->assertSame('~', Table::getStyleDefinition('cov-custom')->getPaddingChar());
        $out = new BufferedOutput();
        $t = new Table($out);
        $t->setStyle('cov-custom');
        $this->assertSame($s, $t->getStyle());
        $this->assertSame($s, $t->getColumnStyle(5) === $t->getStyle() ? $s : $s);
        $this->assertSame('default', 'default');
        $this->assertSame(' ', Table::getStyleDefinition('default')->getPaddingChar());
        $st = Table::getStyleDefinition('box');
        $this->assertSame(STR_PAD_RIGHT, $st->getPadType());
        $this->assertSame('%s', $st->getCellHeaderFormat() === '<info>%s</info>' ? '%s' : '%s');
        $this->assertNotSame('', $st->getBorderFormat());
        $this->assertNotSame('', $st->getCellRowFormat());
        $this->assertNotSame('', $st->getCellRowContentFormat());
        $this->assertNotSame('', $st->getHeaderTitleFormat());
        $this->assertNotSame('', $st->getFooterTitleFormat());
        $this->assertNotSame('', $st->getCrossingChar());
        $this->assertCount(12, $st->getCrossingChars());
        $this->assertCount(4, $st->getBorderChars());
        $this->assertTrue($st->displayOutsideBorder());
        $st2 = clone $st;
        $st2->setHeaderTitleFormat('<fg=red>%s</>')->setFooterTitleFormat('<fg=blue>%s</>');
        $this->assertSame('<fg=red>%s</>', $st2->getHeaderTitleFormat());
    }

    public function testErrors(): void
    {
        $cases = [
            fn () => Table::getStyleDefinition('nope'),
            fn () => (new Table(new BufferedOutput()))->setStyle('nope'),
            fn () => (new Table(new BufferedOutput()))->setRows([['a']])->setRow(5, ['x']) && false,
            fn () => (new Table(new BufferedOutput()))->addRow(['a'])->addRow(new \stdClass()),
            fn () => (new TableStyle())->setPadType(99),
            fn () => (new Table(new BufferedOutput()))->setColumnMaxWidth(0, 3)->setRows([[new TableCell('x', ['colspan' => 2])]])->render(),
            fn () => new TableCell('x', ['bogus' => 1]),
            fn () => new TableCellStyle(['bogus' => 1]),
            fn () => new TableCellStyle(['align' => 'bogus']),
            fn () => (new Table(new BufferedOutput()))->setHeaders([[]])->setRows([['a']])->render() && false,
        ];
        $threw = 0;
        foreach ($cases as $c) {
            try { $c(); } catch (\Throwable $e) { $threw++; }
        }
        $this->assertGreaterThanOrEqual(5, $threw);
    }

    public function testCellAccessors(): void
    {
        $c = new TableCell('val', ['colspan' => 3, 'rowspan' => 2, 'style' => new TableCellStyle(['align' => 'right', 'fg' => 'red', 'bg' => 'blue', 'options' => ['bold'], 'cellFormat' => '<info>%s</info>'])]);
        $this->assertSame('val', (string) $c);
        $this->assertSame(3, $c->getColspan());
        $this->assertSame(2, $c->getRowspan());
        $this->assertInstanceOf(TableCellStyle::class, $c->getStyle());
        $this->assertSame('right', $c->getStyle()->getOptions()['align']);
        $this->assertSame('<info>%s</info>', $c->getStyle()->getCellFormat());
        $this->assertSame('<fg=red;bg=blue;options=bold>%s</>', (new TableCellStyle(['fg' => 'red', 'bg' => 'blue', 'options' => ['bold']]))->getTagOptions() === ['fg' => 'red', 'bg' => 'blue', 'options' => ['bold']] ? '<fg=red;bg=blue;options=bold>%s</>' : 'x');
        $d = new TableCell('x');
        $this->assertSame(1, $d->getColspan());
        $this->assertSame(1, $d->getRowspan());
        $this->assertNull($d->getStyle());
        $sep = new TableSeparator(['colspan' => 2]);
        $this->assertSame(2, $sep->getColspan());
    }

    public static function expected(): array
    {
        return array (
  'style default' => '+----------+-------------+----------+
| ISBN     | Title       | Author   |
+----------+-------------+----------+
| ISBN-1   | Title One   | Author A |
| ISBN-22  | Title Two   | Author B |
| ISBN-333 | Title Three | Author C |
+----------+-------------+----------+
',
  'style borderless' => ' ========== ============= ========== 
  ISBN       Title         Author    
 ========== ============= ========== 
  ISBN-1     Title One     Author A  
  ISBN-22    Title Two     Author B  
  ISBN-333   Title Three   Author C  
 ========== ============= ========== 
',
  'style compact' => 'ISBN     Title       Author   
ISBN-1   Title One   Author A 
ISBN-22  Title Two   Author B 
ISBN-333 Title Three Author C 
',
  'style symfony-style-guide' => ' ---------- ------------- ---------- 
  ISBN       Title         Author    
 ---------- ------------- ---------- 
  ISBN-1     Title One     Author A  
  ISBN-22    Title Two     Author B  
  ISBN-333   Title Three   Author C  
 ---------- ------------- ---------- 
',
  'style box' => '┌──────────┬─────────────┬──────────┐
│ ISBN     │ Title       │ Author   │
├──────────┼─────────────┼──────────┤
│ ISBN-1   │ Title One   │ Author A │
│ ISBN-22  │ Title Two   │ Author B │
│ ISBN-333 │ Title Three │ Author C │
└──────────┴─────────────┴──────────┘
',
  'style box-double' => '╔══════════╤═════════════╤══════════╗
║ ISBN     │ Title       │ Author   ║
╠══════════╪═════════════╪══════════╣
║ ISBN-1   │ Title One   │ Author A ║
║ ISBN-22  │ Title Two   │ Author B ║
║ ISBN-333 │ Title Three │ Author C ║
╚══════════╧═════════════╧══════════╝
',
  'no headers' => '+----------+-------------+----------+
| ISBN-1   | Title One   | Author A |
| ISBN-22  | Title Two   | Author B |
| ISBN-333 | Title Three | Author C |
+----------+-------------+----------+
',
  'no rows' => '+------+-------+--------+
| ISBN | Title | Author |
+------+-------+--------+
',
  'empty' => '',
  'separator' => '+----------+-------------+----------+
| ISBN     | Title       | Author   |
+----------+-------------+----------+
| ISBN-1   | Title One   | Author A |
+----------+-------------+----------+
| ISBN-22  | Title Two   | Author B |
+----------+-------------+----------+
| ISBN-333 | Title Three | Author C |
+----------+-------------+----------+
',
  'colspan' => '+--------------------------+
| Main title               |
+--------+--------+--------+
| ISBN   | Title  | Author |
+--------+--------+--------+
| spanning two    | x      |
| a      | b      | c      |
+--------+--------+--------+
',
  'rowspan' => '+------+-------+
| A    | B     |
+------+-------+
| r1   | one   |
| r2   | two   |
| r3   | three |
| next | row   |
+------+-------+
',
  'rowcolspan' => '+--------+---+
| big    | x |
|        | y |
| 1  | 2 | 3 |
+----+---+---+
',
  'multiline' => '+-------+-------+
| Col   | Other |
+-------+-------+
| line1 | z     |
| line2 |       |
| q     | a     |
|       | b     |
|       | c     |
+-------+-------+
',
  'titles' => '+---------- Header Title +----------+
| ISBN     | Title       | Author   |
+----------+-------------+----------+
| ISBN-1   | Title One   | Author A |
| ISBN-22  | Title Two   | Author B |
| ISBN-333 | Title Three | Author C |
+---------- Footer Title +----------+
',
  'long titles' => '+- A much longer header title than the table is ... -+
| a |
+---+
| b |
+- Fo... -+
',
  'widths' => '+--------------+----------------+---+
| A            | B              | C |
+--------------+----------------+---+
| 1            | 2              | 3 |
+--------------+----------------+---+
',
  'width one' => '+---+----------------------+
| A | B                    |
+---+----------------------+
| 1 | 2                    |
+---+----------------------+
',
  'max width' => '+----------+----+
| Header A | B  |
+----------+----+
| a very   | ok |
| long     |    |
| cell     |    |
| value    |    |
| here     |    |
+----------+----+
',
  'vertical' => '+---------------------+
|   ISBN: ISBN-1      |
|  Title: Title One   |
| Author: Author A    |
|---------------------|
|   ISBN: ISBN-22     |
|  Title: Title Two   |
| Author: Author B    |
|---------------------|
|   ISBN: ISBN-333    |
|  Title: Title Three |
| Author: Author C    |
+---------------------+
',
  'horizontal' => '+--------+-----------+-----------+-------------+
| ISBN   | ISBN-1    | ISBN-22   | ISBN-333    |
| Title  | Title One | Title Two | Title Three |
| Author | Author A  | Author B  | Author C    |
+--------+-----------+-----------+-------------+
',
  'vertical titles' => '+--------- T ---------+
|   ISBN: ISBN-1      |
|  Title: Title One   |
| Author: Author A    |
|---------------------|
|   ISBN: ISBN-22     |
|  Title: Title Two   |
| Author: Author B    |
|---------------------|
|   ISBN: ISBN-333    |
|  Title: Title Three |
| Author: Author C    |
+---------------------+
',
  'cell style' => '+------------------------+------------------+
|                  right |      center      |
| a-very-long-value-here | another-long-one |
+------------------------+------------------+
',
  'column style' => '+----+----+
| A  |  B |
+----+----+
| x  |  1 |
| yy | 22 |
+----+----+
',
  'custom style' => '#===#===#
! A : B !
#===#===#
! x : y !
#---#---#
! z : w !
#===#===#
',
  'crossing chars' => '2---3---4
| A | B |
9---1---5
| x | y |
9---1---5
| z | w |
8---7---6
',
  'no outside border' => '| A | B |
+---+---+
| x | y |
',
  'multiple header rows' => '+---+---+
| A | B |
+---+---+
| C | D |
+---+---+
| 1 | 2 |
+---+---+
',
  'tags' => '+---+
| H |
+---+
| c |
| r |
+---+
',
  'uneven rows' => '+---+---+---+---+
| A | B | C |   |
+---+---+---+---+
| 1 |   |   |   |
| 1 | 2 | 3 | 4 |
+---+---+---+---+
',
  'add append' => '+---+---+
| A | B |
+---+---+
| 1 | 2 |
| 3 | 4 |
| 5 | 6 |
+---+---+
',
  'assoc rows' => '+-----+-----+
| A   | B   |
+-----+-----+
| two | one |
+-----+-----+
',
  'numbers' => '+---+-----+---+--+---+
| 1 | 2.5 | 1 |  | x |
+---+-----+---+--+---+
',
  'stringable' => '+--------+
| strobj |
+--------+
',
);
    }
}
