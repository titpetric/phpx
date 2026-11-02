# Filament

`filament/filament` 4.16.0. `panel.php` is the `PanelProvider` a Laravel
application registers. Filament has no standalone entry: `Panel::make()->id()`
calls `Container::runningInConsole()`, and `php` fails that outside a booted
application. The snippet cannot run alone, and this folder holds no fixture. A
probe needs a Laravel application with the provider registered.

Filament is a Laravel package over Livewire, and its blockers are Laravel's:

- the generated composer platform check requires `PHP_VERSION_ID >= 80401`, and
  phpscript reports `8.4.0`.
- `symfony/polyfill-mbstring/bootstrap80.php` uses `??`, which phpscript
  refuses at parse time.

The Livewire front end is not reached. [The report](../../report.md) keeps the
Laravel blockers and this package apart.
