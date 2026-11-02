<?php

namespace App\Providers;

use Filament\Panel;
use Filament\PanelProvider;

/**
 * The provider a Laravel application registers in bootstrap/providers.php.
 * Filament has no standalone entry: Panel::make()->id() calls
 * Container::runningInConsole(), and php fails that outside a booted
 * application.
 */
class AdminPanelProvider extends PanelProvider
{
	public function panel(Panel $panel): Panel
	{
		return $panel->id('admin')->path('admin');
	}
}
