<?php
use PHPUnit\Framework\TestCase;

final class FilamentEntryTest extends TestCase
{
    public function testPanelClassesAutoload(): void
    {
        $this->assertTrue(class_exists('Filament\Panel'));
        $this->assertTrue(class_exists('Filament\PanelProvider'));
        $this->assertTrue(interface_exists('Filament\Contracts\Plugin'));
    }
    public function testSupportColors(): void
    {
        $this->assertTrue(class_exists('Filament\Support\Colors\Color'));
        $this->assertIsArray(\Filament\Support\Colors\Color::Blue);
        $this->assertArrayHasKey(500, \Filament\Support\Colors\Color::Blue);
    }
}
