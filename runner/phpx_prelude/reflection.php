<?php
class ReflectionException extends Exception {}

class ReflectionAttribute {
    const IS_INSTANCEOF = 2;
    private $name;
    private $args;
    private $file;
    private $class;
    public function __construct($name, $args, $file, $class) {
        $this->name = $name;
        $this->args = $args;
        $this->file = $file;
        $this->class = $class;
    }
    public function getName() { return $this->name; }
    public function getArguments() { return __phpx_attribute_args($this->file, $this->class, $this->args); }
    public function newInstance() { $n = $this->name; return new $n(...$this->getArguments()); }
    public function getTarget() { return 1; }
    public function isRepeated() { return false; }
}

function __phpx_attributes_of($file, $key, $class, $name, $flags) {
    $out = [];
    if (!is_string($file)) { return $out; }
    foreach (__phpx_attributes($file, $key) as $pair) {
        $match = $name === null || strcasecmp(ltrim($name, '\\'), $pair[0]) === 0
            || ((($flags & 2) !== 0) && is_a($pair[0], ltrim($name, '\\'), true));
        if ($match) { $out[] = new ReflectionAttribute($pair[0], $pair[1], $file, $class); }
    }
    return $out;
}

class ReflectionNamedType {
    private $type;
    public function __construct($type) { $this->type = $type; }
    public function getName() { return ltrim($this->type, '?'); }
    public function allowsNull() {
        return $this->type[0] === '?' || $this->type === 'mixed' || $this->type === 'null'
            || stripos('|' . $this->type . '|', '|null|') !== false;
    }
    public function isBuiltin() {
        return in_array(strtolower(ltrim($this->type, '?')), ['int', 'float', 'string', 'bool', 'array', 'mixed',
            'void', 'null', 'callable', 'iterable', 'object', 'never', 'false', 'true', 'static', 'self'], true);
    }
    public function __toString() { return $this->type; }
}

class ReflectionParameter {
    public string $name = '';
    private $info;
    private $owner;
    public function __construct($info, $owner = null) {
        $this->info = $info;
        $this->owner = $owner;
        $this->name = $info['name'];
    }
    public function getName() { return $this->name; }
    public function getPosition() { return $this->info['position']; }
    public function hasType() { return $this->info['type'] !== null; }
    public function getType() { return $this->info['type'] === null ? null : new ReflectionNamedType($this->info['type']); }
    public function allowsNull() { return $this->info['type'] === null || (new ReflectionNamedType($this->info['type']))->allowsNull(); }
    public function isOptional() { return $this->info['hasDefault'] || $this->info['variadic']; }
    public function isDefaultValueAvailable() { return $this->info['hasDefault']; }
    public function getDefaultValue() {
        return __phpx_param_default($this->owner[0], $this->owner[1], $this->info['position']);
    }
    public function getClosureThis() { return null; }
    public function getClosureScopeClass() { return null; }
    public function getDeclaringClass() { return $this->owner[1] === '' ? null : new ReflectionClass($this->owner[0]); }
    public function isVariadic() { return $this->info['variadic']; }
    public function isPassedByReference() { return $this->info['byRef']; }
    public function isPromoted() { return $this->info['promoted']; }
    public function getAttributes($name = null, $flags = 0) { return []; }
    public function getDeclaringFunction() { return $this->owner[1] === '' ? new ReflectionFunction($this->owner[2]) : new ReflectionMethod($this->owner[0], $this->owner[1]); }
}

abstract class ReflectionFunctionAbstract {
    public string $name = '';
    protected $info;
    public function getName() { return $this->name; }
    public function getShortName() { $p = strrpos($this->name, '\\'); return $p === false ? $this->name : substr($this->name, $p + 1); }
    public function getParameters() {
        $out = [];
        foreach ($this->info['params'] as $p) { $out[] = new ReflectionParameter($p, $this->owner()); }
        return $out;
    }
    public function getNumberOfParameters() { return count($this->info['params']); }
    public function getNumberOfRequiredParameters() {
        $n = 0;
        foreach ($this->info['params'] as $p) { if (!$p['hasDefault'] && !$p['variadic']) { $n++; } }
        return $n;
    }
    public function hasReturnType() { return $this->info['returnType'] !== null; }
    public function getReturnType() { return $this->info['returnType'] === null ? null : new ReflectionNamedType($this->info['returnType']); }
    public function getDocComment() { return false; }
    public function getAttributes($name = null, $flags = 0) { return []; }
    public function getFileName() { return $this->info['file']; }
    public function getStartLine() { return 0; }
    public function getEndLine() { return 0; }
    public function isDeprecated() { return false; }
    public function getExtension() { return null; }
    public function getExtensionName() { return false; }
    public function getStaticVariables() { return []; }
    public function isVariadic() { foreach ($this->info['params'] as $p) { if ($p['variadic']) { return true; } } return false; }
    public function isInternal() { return false; }
    public function isUserDefined() { return true; }
    public function isGenerator() { return false; }
    public function returnsReference() { return false; }
    abstract protected function owner();
}

class ReflectionMethod extends ReflectionFunctionAbstract {
    const IS_STATIC = 16;
    const IS_PUBLIC = 1;
    const IS_PROTECTED = 2;
    const IS_PRIVATE = 4;
    const IS_ABSTRACT = 64;
    const IS_FINAL = 32;
    public string $class = '';
    public function __construct($objectOrMethod, $method = null) {
        if ($method === null && is_string($objectOrMethod) && strpos($objectOrMethod, '::') !== false) {
            [$objectOrMethod, $method] = explode('::', $objectOrMethod, 2);
        }
        $cls = __phpx_class_info($objectOrMethod);
        if ($cls === false) {
            throw new ReflectionException('Class "' . (is_object($objectOrMethod) ? get_class($objectOrMethod) : $objectOrMethod) . '" does not exist');
        }
        foreach ($cls['methods'] as $name => $m) {
            if (strtolower($name) === strtolower($method)) {
                $this->info = $m;
                $this->name = $m['name'];
                $this->class = $m['class'];
                return;
            }
        }
        throw new ReflectionException('Method ' . $cls['name'] . '::' . $method . '() does not exist');
    }
    protected function owner() { return [$this->class, $this->name, '']; }
    public function getDocComment() { return is_string($this->info['file']) ? __phpx_doc($this->info['file'], $this->class . '::' . $this->name) : false; }
    public function getAttributes($name = null, $flags = 0) { return __phpx_attributes_of($this->info['file'], $this->class . '::' . $this->name, $this->class, $name, $flags); }
    public function isPublic() { return $this->info['visibility'] === 'public'; }
    public function isProtected() { return $this->info['visibility'] === 'protected'; }
    public function isPrivate() { return $this->info['visibility'] === 'private'; }
    public function isStatic() { return $this->info['static']; }
    public function isAbstract() { return $this->info['abstract']; }
    public function isFinal() { return false; }
    public function isConstructor() { return strtolower($this->name) === '__construct'; }
    public function isDestructor() { return strtolower($this->name) === '__destruct'; }
    public function getDeclaringClass() { return new ReflectionClass($this->class); }
    public function setAccessible($accessible) {}
    public function getModifiers() {
        $m = $this->info['visibility'] === 'public' ? 1 : ($this->info['visibility'] === 'protected' ? 2 : 4);
        return $m | ($this->info['static'] ? 16 : 0) | ($this->info['abstract'] ? 64 : 0);
    }
    public function invoke($object = null, ...$args) { return $this->invokeArgs($object, $args); }
    public function invokeArgs($object = null, array $args = []) {
        if ($this->info['static'] || $object === null) { return call_user_func_array([$this->class, $this->name], $args); }
        return call_user_func_array([$object, $this->name], $args);
    }
    public function getClosure($object = null) {
        $self = $this;
        return function (...$args) use ($self, $object) { return $self->invokeArgs($object, $args); };
    }
}

class ReflectionFunction extends ReflectionFunctionAbstract {
    private $callable;
    public function __construct($function) {
        $this->callable = $function;
        $info = __phpx_func_info($function);
        if ($info === false) {
            throw new ReflectionException('Function ' . (is_string($function) ? $function : 'closure:' . get_debug_type($function)) . '() does not exist');
        }
        $this->info = $info;
        $this->name = $info['name'];
    }
    protected function owner() { return ['', '', $this->callable]; }
    public function invoke(...$args) { return call_user_func_array($this->callable, $args); }
    public function invokeArgs(array $args = []) { return call_user_func_array($this->callable, $args); }
    public function isClosure() { return $this->callable instanceof Closure; }
    public function getClosureThis() { return null; }
    public function getClosureCalledClass() { return null; }
    public function getClosureScopeClass() { return null; }
    public function getClosure() { return $this->callable; }
}

class ReflectionProperty {
    const IS_STATIC = 16;
    const IS_READONLY = 128;
    const IS_PUBLIC = 1;
    const IS_PROTECTED = 2;
    const IS_PRIVATE = 4;
    public string $name = '';
    public string $class = '';
    private $info;
    public function __construct($class, $property) {
        $cls = __phpx_class_info($class);
        if ($cls === false || !isset($cls['props'][$property])) {
            throw new ReflectionException('Property ' . (is_object($class) ? get_class($class) : $class) . '::$' . $property . ' does not exist');
        }
        $this->info = $cls['props'][$property];
        $this->name = $property;
        $this->class = $this->info['class'];
        $owner = __phpx_class_info($this->class);
        $this->file = $owner === false ? false : $owner['file'];
    }
    private $file = false;
    public function getName() { return $this->name; }
    public function isPublic() { return $this->info['visibility'] === 'public'; }
    public function isProtected() { return $this->info['visibility'] === 'protected'; }
    public function isPrivate() { return $this->info['visibility'] === 'private'; }
    public function isStatic() { return $this->info['static']; }
    public function isReadOnly() { return $this->info['readonly']; }
    public function isPromoted() { return false; }
    public function hasType() { return $this->info['type'] !== null; }
    public function getType() { return $this->info['type'] === null ? null : new ReflectionNamedType($this->info['type']); }
    public function hasDefaultValue() { return $this->info['hasDefault']; }
    public function getDeclaringClass() { return new ReflectionClass($this->class); }
    public function getDocComment() { return is_string($this->file) ? __phpx_doc($this->file, $this->class . '::$' . $this->name) : false; }
    public function getAttributes($name = null, $flags = 0) { return __phpx_attributes_of($this->file, $this->class . '::$' . $this->name, $this->class, $name, $flags); }
    public function setAccessible($accessible) {}
    public function getValue($object = null) {
        if ($this->info['static']) { return __phpx_static_prop($this->class, $this->name); }
        return __phpx_prop_get($object, $this->name);
    }
    public function setValue($objectOrValue = null, $value = null) {
        if ($this->info['static']) { __phpx_static_prop($this->class, $this->name, $value ?? $objectOrValue, true); return; }
        __phpx_prop_set($objectOrValue, $this->name, $value);
    }
    public function isInitialized($object = null) { return __phpx_prop_isset($object, $this->name); }
}

class ReflectionClass {
    const IS_IMPLICIT_ABSTRACT = 16;
    const IS_EXPLICIT_ABSTRACT = 64;
    const IS_FINAL = 32;
    public string $name = '';
    protected $info;
    public function __construct($objectOrClass) {
        $info = __phpx_class_info($objectOrClass);
        if ($info === false) {
            throw new ReflectionException('Class "' . (is_object($objectOrClass) ? get_class($objectOrClass) : $objectOrClass) . '" does not exist');
        }
        $this->info = $info;
        $this->name = $info['name'];
    }
    public function getName() { return $this->name; }
    public function getShortName() { return $this->info['short']; }
    public function getNamespaceName() { return $this->info['namespace']; }
    public function inNamespace() { return $this->info['namespace'] !== ''; }
    public function isAbstract() { return $this->info['abstract'] && !$this->info['interface']; }
    public function isFinal() { return $this->info['final']; }
    public function isReadOnly() { return $this->info['readonly'] ?? false; }
    public function isInterface() { return $this->info['interface']; }
    public function isTrait() { return $this->info['trait'] ?? false; }
    public function isEnum() { return false; }
    public function isAnonymous() { return strpos($this->name, 'class@anonymous') === 0; }
    public function isInternal() { return $this->info['internal'] ?? false; }
    public function isUserDefined() { return !($this->info['internal'] ?? false); }
    public function isInstantiable() {
        if ($this->info['interface'] || $this->info['abstract']) { return false; }
        foreach ($this->info['methods'] as $m) {
            if (strtolower($m['name']) === '__construct') { return $m['visibility'] === 'public'; }
        }
        return true;
    }
    public function isCloneable() { return $this->isInstantiable(); }
    public function isInstance($object) { return $object instanceof $this->name; }
    public function isSubclassOf($class) {
        $class = $class instanceof ReflectionClass ? $class->getName() : ltrim($class, '\\');
        if (strtolower($class) === strtolower($this->name)) { return false; }
        return is_a($this->name, $class, true);
    }
    public function implementsInterface($interface) {
        $interface = $interface instanceof ReflectionClass ? $interface->getName() : ltrim($interface, '\\');
        foreach ($this->info['interfaces'] as $i) { if (strtolower($i) === strtolower($interface)) { return true; } }
        return false;
    }
    public function getFileName() { return $this->info['file']; }
    public function getStartLine() { return 0; }
    public function getEndLine() { return 0; }
    public function getExtension() { return null; }
    public function getExtensionName() { return false; }
    public function isIterable() { return $this->implementsInterface('Traversable'); }
    public function isIterateable() { return $this->isIterable(); }
    public function getDocComment() { return is_string($this->info['file']) ? __phpx_doc($this->info['file'], $this->name) : false; }
    public function getAttributes($name = null, $flags = 0) { return __phpx_attributes_of($this->info['file'], $this->name, $this->name, $name, $flags); }
    public function getParentClass() { return $this->info['parent'] === null ? false : new ReflectionClass($this->info['parent']); }
    public function getInterfaceNames() { return $this->info['interfaces']; }
    public function getInterfaces() {
        $out = [];
        foreach ($this->info['interfaces'] as $i) { $out[$i] = new ReflectionClass($i); }
        return $out;
    }
    public function getTraitNames() { return $this->info['traits'] ?? []; }
    public function getTraits() {
        $out = [];
        foreach ($this->getTraitNames() as $t) { $out[$t] = new ReflectionClass($t); }
        return $out;
    }
    public function hasMethod($name) {
        foreach ($this->info['methods'] as $m) { if (strtolower($m['name']) === strtolower($name)) { return true; } }
        return false;
    }
    public function getMethod($name) { return new ReflectionMethod($this->name, $name); }
    public function getMethods($filter = null) {
        $out = [];
        foreach ($this->info['methods'] as $m) {
            $r = new ReflectionMethod($this->name, $m['name']);
            if ($filter === null || ($r->getModifiers() & $filter) !== 0) { $out[] = $r; }
        }
        return $out;
    }
    public function getConstructor() { return $this->hasMethod('__construct') ? new ReflectionMethod($this->name, '__construct') : null; }
    public function hasProperty($name) { return isset($this->info['props'][$name]); }
    public function getProperty($name) { return new ReflectionProperty($this->name, $name); }
    public function getProperties($filter = null) {
        $out = [];
        foreach ($this->info['props'] as $n => $p) {
            $vis = $p['visibility'] === 'public' ? 1 : ($p['visibility'] === 'protected' ? 2 : 4);
            if ($filter === null || (($vis | ($p['static'] ? 16 : 0)) & $filter) !== 0) { $out[] = new ReflectionProperty($this->name, $n); }
        }
        return $out;
    }
    public function getStaticPropertyValue($name, $default = null) { return __phpx_static_prop($this->name, $name); }
    public function setStaticPropertyValue($name, $value) { __phpx_static_prop($this->name, $name, $value, true); }
    public function hasConstant($name) { return isset($this->info['consts'][$name]); }
    public function getConstants($filter = null) {
        $out = [];
        foreach ($this->info['consts'] as $n => $_) { $out[$n] = constant($this->name . '::' . $n); }
        return $out;
    }
    public function getConstant($name) { return isset($this->info['consts'][$name]) ? constant($this->name . '::' . $name) : false; }
    public function newInstance(...$args) { $n = $this->name; return new $n(...$args); }
    public function newInstanceArgs(array $args = []) { $n = $this->name; return new $n(...$args); }
    public function newInstanceWithoutConstructor() { return __phpx_new_noctor($this->name); }
    public function getModifiers() { return 0; }
}

class ReflectionObject extends ReflectionClass {}
