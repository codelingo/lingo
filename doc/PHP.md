# PHP

## Syntax Facts

Currently PHP facts are not translated into common (un-namespaced) facts, so they must be fully qualified and use the name assigned by the PHP parser. For example, a variable assignment inside a function would be

```yaml
php.stmt_function:
  php.expr_assign
```

To find the full list of supported nodes, see https://github.com/nikic/PHP-Parser/tree/master/lib/PhpParser/Node - the fact name can be derived by replacing `\` with `_` in the qualified class name of a node relative to the PhpParser\Node, all in lower case. For example, https://github.com/nikic/PHP-Parser/blob/master/lib/PhpParser/Node/Scalar/MagicConst/File.php would become `scalar_magicconst_file`. Used in a tenet, it will need the php namespace prefix, ie. `php.scalar_magicconst_file`.

## Properties

Property inspection on PHP facts is unavailable in the current release. Once implemented they will take the common form but be available on either common or php specific facts, ie. the following will be equivalent:

```yaml
match:
  var:
    type: "string"
```

```
match:
  php.expr_variable:
    type: "string"
```
