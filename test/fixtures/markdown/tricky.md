# A heading

Some _emphasis_ and *also emphasis*, __strong__ and **strong**, and `code`.
A line that wraps
onto the next line without a blank line.

Hard break with two spaces  
and a backslash\
too.

* star list
* second item
  continued

- dash list
- [ ] a task
- [x] done task

1. one
2. two
10. ten

1) paren list

> A quote
> over two lines

---

***

Setext heading
==============

Another
-------

[inline link](https://example.com "title") and [reference link][ref] and <https://auto.example>.

![an image](/images/loaf.jpg)

[ref]: https://example.com/ref

```go
func main() {}
```

~~~
tilde fence
~~~

    indented code

| a | b |
|---|:-:|
| 1 | 2 |

<div class="raw">Some HTML</div>

Escapes: \* not a list, 3 \< 4, a_b_c, 100%.

{{< figure src="/images/loaf.jpg" caption="Our loaf" alt="A loaf" >}}

{{% notice info %}}
This is **inside** a shortcode.
{{% /notice %}}

Inline {{< icon "bread" >}} shortcode in text.

Footnote here[^1].

[^1]: The footnote.

Trailing text with no final newline
