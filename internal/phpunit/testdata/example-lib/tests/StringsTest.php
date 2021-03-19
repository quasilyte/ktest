<?php

use PHPUnit\Framework\TestCase;
use ExampleLib\Strings;

class StringsTest extends TestCase {
    public function testContains() {
        $this->assertSame(Strings::contains('foo', 'bar'), false);
        $this->assertSame(Strings::contains('foo', 'foo'), true);
    }

    public function testHasPrefix() {
        $this->assertSame(Strings::hasPrefix('Hello World', 'Hello'), true);
        $this->assertSame(Strings::hasPrefix('Hello World', 'ello'), false);
    }
}
