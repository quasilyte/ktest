## Overview

`ktest` is a project that makes [kphp](https://github.com/VKCOM/kphp/) programs easier to test.

* `kphp phpunit` can run [PHPUnit](https://github.com/sebastianbergmann/phpunit) tests
* `kphp bench` runs benchmark tests (not implemented yet)

## Example

Imagine that we have an ordinary `PHPUnit` test:

```php
<?php

use PHPUnit\Framework\TestCase;
use ExampleLib\Strings;

class StringsTest extends TestCase {
    public function testContains() {
        $this->assertSame(Strings::contains('foo', 'bar'), false);
        $this->assertTrue(Strings::contains('foo', 'foo'));
    }

    public function testHasPrefix() {
        $this->assertSame(Strings::hasPrefix('Hello World', 'Hello'), true);
        $this->assertFalse(Strings::hasPrefix('Hello World', 'ello'));
    }
}
```

It comes without a surprise that you can run it with `phpunit` tool:

```
$ ./vendor/bin/phpunit tests

......                                                              6 / 6 (100%)

Time: 70 ms, Memory: 4.00 MB

OK (6 tests, 14 assertions)
```

When you're using `phpunit`, tests are executed as PHP, not KPHP.

`ktest` makes it possible to run your phpunit-compatible tests with KPHP:

```
$ ktest phpunit tests

.... 4 / 6 (66%) OK
.. 6 / 6 (100%) OK

Time: 10.74657386s

OK (6 tests, 14 assertions)
```

## TODO

* Mocks
* Benchmarks

## Limitations

* Assert functions can't be used for objects (class instances)
* No custom comparators for assert functions
