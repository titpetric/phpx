<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Input\{ArrayInput, InputInterface};
use Symfony\Component\Console\Output\{BufferedOutput, OutputInterface};
use Symfony\Component\Console\Style\SymfonyStyle;
use Symfony\Component\Console\Helper\{TableSeparator, TableCell};
use Symfony\Component\Console\Question\{Question, ChoiceQuestion};

final class ConsoleCov6Test extends TestCase
{

    /** @return array<string, callable():string> */
    public static function cases(): array
    {
        $mk = function (callable $f, string $in = '', ?int $verb = null): callable {
            return function () use ($f, $in, $verb): string {
                $fp = fopen('php://memory', 'r+');
                fwrite($fp, $in);
                rewind($fp);
                $input = new ArrayInput([]);
                $input->setStream($fp);
                $out = new BufferedOutput($verb ?? OutputInterface::VERBOSITY_NORMAL);
                $out->setDecorated(false);
                putenv('COLUMNS=60');
                $io = new SymfonyStyle($input, $out);
                $f($io);
                return $out->fetch();
            };
        };
        $c = [];
        $c['title'] = $mk(fn (SymfonyStyle $io) => $io->title('The Title'));
        $c['section'] = $mk(fn (SymfonyStyle $io) => $io->section('A Section'));
        $c['listing'] = $mk(fn (SymfonyStyle $io) => $io->listing(['one', 'two', 'three']));
        $c['text'] = $mk(function (SymfonyStyle $io) { $io->text('single'); $io->text(['multi', 'ple']); $io->comment('a comment'); $io->comment(['c1', 'c2']); });
        $c['success'] = $mk(function (SymfonyStyle $io) { $io->success('Done'); $io->success(['Two', 'lines']); });
        $c['error'] = $mk(function (SymfonyStyle $io) { $io->error('Bad'); $io->error(['Two', 'lines']); });
        $c['warning'] = $mk(function (SymfonyStyle $io) { $io->warning('Careful'); $io->warning(['W1', 'W2']); });
        $c['note'] = $mk(function (SymfonyStyle $io) { $io->note('Noted'); $io->note(['N1', 'N2']); });
        $c['info'] = $mk(function (SymfonyStyle $io) { $io->info('Information'); $io->info(['I1', 'I2']); });
        $c['caution'] = $mk(function (SymfonyStyle $io) { $io->caution('Caution!'); $io->caution(['C1', 'C2']); });
        $c['block'] = $mk(function (SymfonyStyle $io) {
            $io->block('plain block');
            $io->block('typed', 'TYPE', 'fg=white;bg=blue', ' > ', true);
            $io->block(['m1', 'm2'], 'T', 'info', '# ', false, false);
            $io->block('a very long block message that is certainly going to need wrapping when the terminal is narrow so we can see it', 'LONG', 'comment', ' ', true);
        });
        $c['table'] = $mk(function (SymfonyStyle $io) { $io->table(['A', 'B'], [['1', '2'], ['3', '4']]); $io->horizontalTable(['A', 'B'], [['1', '2'], ['3', '4']]); });
        $c['definition list'] = $mk(fn (SymfonyStyle $io) => $io->definitionList('Heading', ['key1' => 'val1'], new TableSeparator(), ['key2' => 'val2'], ['key3' => 'val3']));
        $c['newline writeln'] = $mk(function (SymfonyStyle $io) { $io->newLine(); $io->newLine(2); $io->writeln('a'); $io->writeln(['b', 'c']); $io->write('d'); $io->write(['e', 'f'], true); $io->write('g'); $io->newLine(); });
        $c['ask'] = $mk(function (SymfonyStyle $io) { $r = $io->ask('Name', 'def'); $io->writeln("got:$r"); $r = $io->ask('Name', null, fn ($v) => strtoupper($v)); $io->writeln("got:$r"); $r = $io->ask('Name', 'def'); $io->writeln("got:$r"); }, "bob\nal\n\n");
        $c['confirm'] = $mk(function (SymfonyStyle $io) { $io->writeln($io->confirm('Sure?') ? 'yes' : 'no'); $io->writeln($io->confirm('Sure?', false) ? 'yes' : 'no'); $io->writeln($io->confirm('Sure?', false) ? 'yes' : 'no'); }, "n\ny\n\n");
        $c['choice'] = $mk(function (SymfonyStyle $io) { $io->writeln($io->choice('Pick', ['a', 'b', 'c'], 'b')); $io->writeln(implode('+', $io->choice('Multi', ['a', 'b', 'c'], null, true))); $io->writeln((string) $io->choice('Pick', ['k' => 'v', 'w' => 'x'])); }, "\n0,2\nw\n");
        $c['askQuestion'] = $mk(fn (SymfonyStyle $io) => $io->writeln($io->askQuestion(new Question('Free', 'x'))), "typed\n");
        $c['progress'] = $mk(function (SymfonyStyle $io) { $io->progressStart(3); $io->progressAdvance(); $io->progressAdvance(2); $io->progressFinish(); });
        $c['progress nomax'] = $mk(function (SymfonyStyle $io) { $io->progressStart(); $io->progressAdvance(); $io->progressFinish(); });
        $c['progress iterate'] = $mk(function (SymfonyStyle $io) { $n = 0; foreach ($io->progressIterate([1, 2, 3]) as $v) { $n += $v; } $io->writeln("n=$n"); });
        $c['createProgressBar'] = $mk(function (SymfonyStyle $io) { $b = $io->createProgressBar(2); $b->start(); $b->advance(); $b->finish(); });
        $c['tree'] = $mk(function (SymfonyStyle $io) { $io->tree(['a' => ['a1', 'a2'], 'b'], 'root'); });
        $c['error style'] = $mk(function (SymfonyStyle $io) { $e = $io->getErrorStyle(); $e->writeln('to-error'); $io->writeln($e instanceof SymfonyStyle ? 'ok' : 'no'); $io->createTable()->setRows([['x']])->render(); });
        $c['verbosity'] = $mk(function (SymfonyStyle $io) { $io->writeln('v', OutputInterface::VERBOSITY_VERBOSE); $io->writeln('n', OutputInterface::VERBOSITY_NORMAL); $io->writeln('d', OutputInterface::VERBOSITY_DEBUG); $io->writeln($io->isVerbose() ? 'isv' : 'notv'); $io->writeln($io->isVeryVerbose() ? 'isvv' : 'notvv'); $io->writeln($io->isDebug() ? 'isd' : 'notd'); $io->writeln($io->isQuiet() ? 'isq' : 'notq'); }, '', OutputInterface::VERBOSITY_VERBOSE);
        $c['mixed sequence'] = $mk(function (SymfonyStyle $io) { $io->title('T'); $io->text('t1'); $io->section('S'); $io->listing(['a']); $io->success('ok'); $io->table(['h'], [['r']]); $io->writeln('end'); });
        $c['formatted'] = $mk(function (SymfonyStyle $io) { $io->success('<info>tag</info> text'); $io->text('<comment>c</comment>'); $io->title('<error>t</error>'); });
        return $c;
    }

    // BEGIN-GEN
    private function runCase(string $name): void
    {
        $expected = self::expected();
        $cases = self::cases();
        $this->assertSame($expected[$name], $cases[$name](), $name);
    }

    public function testCase00(): void { $this->runCase('title'); }
    public function testCase01(): void { $this->runCase('section'); }
    public function testCase02(): void { $this->runCase('listing'); }
    public function testCase03(): void { $this->runCase('text'); }
    public function testCase04(): void { $this->runCase('success'); }
    public function testCase05(): void { $this->runCase('error'); }
    public function testCase06(): void { $this->runCase('warning'); }
    public function testCase07(): void { $this->runCase('note'); }
    public function testCase08(): void { $this->runCase('info'); }
    public function testCase09(): void { $this->runCase('caution'); }
    public function testCase10(): void { $this->runCase('block'); }
    public function testCase11(): void { $this->runCase('table'); }
    public function testCase12(): void { $this->runCase('definition list'); }
    public function testCase13(): void { $this->runCase('newline writeln'); }
    public function testCase14(): void { $this->runCase('ask'); }
    public function testCase15(): void { $this->runCase('confirm'); }
    public function testCase16(): void { $this->runCase('choice'); }
    public function testCase17(): void { $this->runCase('askQuestion'); }
    public function testCase18(): void { $this->runCase('progress'); }
    public function testCase19(): void { $this->runCase('progress nomax'); }
    public function testCase20(): void { $this->runCase('progress iterate'); }
    public function testCase21(): void { $this->runCase('createProgressBar'); }
    public function testCase22(): void { $this->runCase('tree'); }
    public function testCase23(): void { $this->runCase('error style'); }
    public function testCase24(): void { $this->runCase('verbosity'); }
    public function testCase25(): void { $this->runCase('mixed sequence'); }
    public function testCase26(): void { $this->runCase('formatted'); }

    public static function expected(): array
    {
        return array (
  'title' => '
The Title
=========

',
  'section' => '
A Section
---------

',
  'listing' => ' * one
 * two
 * three

',
  'text' => ' single
 multi
 ple

 // a comment                                               

 // c1                                                      
 //                                                         
 // c2                                                      

',
  'success' => '
 [OK] Done                                                  

 [OK] Two                                                   
                                                            
      lines                                                 

',
  'error' => '
 [ERROR] Bad                                                

 [ERROR] Two                                                
                                                            
         lines                                              

',
  'warning' => '
 [WARNING] Careful                                          

 [WARNING] W1                                               
                                                            
           W2                                               

',
  'note' => '
 ! [NOTE] Noted                                             

 ! [NOTE] N1                                                
 !                                                          
 !        N2                                                

',
  'info' => '
 [INFO] Information                                         

 [INFO] I1                                                  
                                                            
        I2                                                  

',
  'caution' => '
 ! [CAUTION] Caution!                                       

 ! [CAUTION] C1                                             
 !                                                          
 !           C2                                             

',
  'block' => '
 plain block                                                

 > [TYPE] typed                                             

# [T] m1                                                    
#                                                           
#     m2                                                    

 [LONG] a very long block message that is certainly going to
        need wrapping when the terminal is narrow so we can 
        see it                                              

',
  'table' => ' --- --- 
  A   B  
 --- --- 
  1   2  
  3   4  
 --- --- 

 --- --- --- 
  A   1   3  
  B   2   4  
 --- --- --- 

',
  'definition list' => ' ------------- 
  Heading      
  key1   val1  
 ------ ------ 
  key2   val2  
  key3   val3  
 ------ ------ 

',
  'newline writeln' => '


a
b
c
de
f
g
',
  'ask' => '
 Name [def]:
 > 
got:bob

 Name:
 > 
got:AL

 Name [def]:
 > 
got:def
',
  'confirm' => '
 Sure? (yes/no) [yes]:
 > 
no

 Sure? (yes/no) [no]:
 > 
yes

 Sure? (yes/no) [no]:
 > 
no
',
  'choice' => '
 Pick [b]:
  [0] a
  [1] b
  [2] c
 > 
b

 Multi:
  [0] a
  [1] b
  [2] c
 > 
a+c

 Pick:
  [k] v
  [w] x
 > 
w
',
  'askQuestion' => '
 Free [x]:
 > 
typed
',
  'progress' => ' 0/3 [░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0%
 3/3 [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] 100%

',
  'progress nomax' => '    0 [░░░░░░░░░░░░░░░░░░░░░░░░░░░░]

',
  'progress iterate' => ' 0/3 [░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0%
 3/3 [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] 100%

n=6
',
  'createProgressBar' => ' 0/2 [░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   0%
 2/2 [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] 100%',
  'tree' => 'root
├── a
│   ├── a1
│   └── a2
└── b
',
  'error style' => 'to-error
ok
 --- 
  x  
 --- 
',
  'verbosity' => 'v
n
isv
notvv
notd
notq
',
  'mixed sequence' => '
T
=

 t1

S
-

 * a

 [OK] ok                                                    

 --- 
  h  
 --- 
  r  
 --- 

end
',
  'formatted' => '
 [OK] <info>tag</info> text                                 

 c

t
=

',
);
    }
    // END-GEN
}
