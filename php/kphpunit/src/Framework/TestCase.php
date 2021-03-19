<?php

namespace KPHPUnit\Framework;

class TestCase {
    /**
     * @param mixed $expected
     * @param mixed $actual
     * @param string $message
     */
    public function assertSame($expected, $actual, string $message = '') {
        if (TestCase::checkIdentical($expected, $actual)) {
            TestCase::ok();
        } else {
            TestCase::fail('SAME', $expected, $actual, $message);
        }
    }

    /**
     * @param mixed $expected
     * @param mixed $actual
     */
    private static function checkIdentical($expected, $actual): bool {
        $float_cmp = is_float($expected) && is_float($actual) &&
                     !is_infinite($expected) && !is_infinite($actual) &&
                     !is_nan($expected) && !is_nan($actual);
        if ($float_cmp) {
            return abs($expected - $actual) < self::EPSILON;
        }
        return $expected === $actual;
    }

    private static function ok() {
        echo json_encode(['ASSERT_OK']) . "\n";
    }

    /**
     * @param string $kind
     * @param mixed $expected
     * @param mixed $actual
     * @param string $message
     */
    private static function fail(string $kind, $expected, $actual, string $message) {
        echo json_encode(["ASSERT_{$kind}_FAILED", $expected, $actual, $message]) . "\n";
        throw new AssertionFailedException();
    }

    private const EPSILON = 0.0000000001;
}
