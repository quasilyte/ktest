![](docs/logo_small.png)

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

Now let's do something more exciting.

Take a look at this `Integers::getFirst` method:

```php
<?php

namespace Foo\Bar;

class Integers {
    /** @param int[] $xs */
    public static function getFirst(array $xs) {
        return $xs[0];
    }
}
```

It's intended to return the first int array item, or `null`, if index 0 is unset.

We can write a test for this method:

```php
<?php

use PHPUnit\Framework\TestCase;
use Foo\Bar\Integers;

class IntegersTest extends TestCase {
    public function testGetFirst() {
        $this->assertSame(Integers::getFirst([]), null);
        $this->assertSame(Integers::getFirst([1]), 1);
    }
}
```

All tests are passing:

```
.                                                                   1 / 1 (100%)

Time: 36 ms, Memory: 4.00 MB

OK (1 test, 2 assertions)
```

But if you try to run it with `ktest`, you'll see how that code would behave in KPHP:

```
F 1 / 1 (100%) FAIL

Time: 4.59874429s

There was 1 failure:

1) IntegersTest::testGetFirst
Failed asserting that null is identical to 0.

IntegersTest.php:8

FAILURES!
Tests: 1, Assertions: 1, Failures: 1.
```

Accessing unset array index can yield a "zero value" instead of null.

Running with `ktest` makes it easier to ensure that your code behaves identically in both PHP and KPHP.

## TODO

* Mocks
* Benchmarks

## Limitations

* Assert functions can't be used for objects (class instances)
* No custom comparators for assert functions
