<?php

namespace ExampleLib;

class Arrays {
    public static function isList(array $array): bool {
        return $array === [] || array_keys($array) === range(0, count($array) - 1);
    }

    public static function isAssoc(array $array): bool {
        return !self::isList($array);
    }

    /**
     * @param array[] $array
     * @return array
     */
    public static function flatten($array) {
        $result = [];
        foreach ($array as $child) {
            $result = array_merge($result, $child);
        }
        return $result;
    }
}
