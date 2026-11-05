<?php
use PHPUnit\Framework\TestCase;
use Symfony\Component\EventDispatcher\EventDispatcher;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Contracts\EventDispatcher\Event;

final class Ping extends Event { public array $log = []; }
final class Sub implements EventSubscriberInterface {
    public static function getSubscribedEvents(): array { return [Ping::class => [['first', 10], ['last', -10]]]; }
    public function first(Ping $e): void { $e->log[] = 'first'; }
    public function last(Ping $e): void { $e->log[] = 'last'; }
}

final class DispatcherEntryTest extends TestCase
{
    public function testPriorityOrder(): void
    {
        $d = new EventDispatcher();
        $d->addListener(Ping::class, function (Ping $e) { $e->log[] = 'mid'; }, 0);
        $d->addSubscriber(new Sub());
        $e = $d->dispatch(new Ping());
        $this->assertSame(['first', 'mid', 'last'], $e->log);
    }
    public function testStopPropagation(): void
    {
        $d = new EventDispatcher();
        $d->addListener('x', function ($e) { $e->stopPropagation(); }, 5);
        $d->addListener('x', function ($e) { throw new \Exception('no'); }, 0);
        $e = $d->dispatch(new Event(), 'x');
        $this->assertTrue($e->isPropagationStopped());
    }
}
