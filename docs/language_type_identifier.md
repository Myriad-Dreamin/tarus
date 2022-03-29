## Language Type Representation

The language type string used by judge service inherits
the format of [MIME-Type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types) Identifier for
classifying
the code written by user.
For exmaple, you can use `application/javascript; standard=es-next` to tell judge service how to process a text input.

There is also another benefit (just sounds interesting) to Language Type Identifier inheriting from MIME-Type. Judge
service can utilize [MIME Sniffing](https://mimesniff.spec.whatwg.org/) technique to automatically correct a wrong
language tagging. It will make a reesponse:

> Oops. You're uploading a file with language type __language/c++__. However, you submit it with __language/python__.

#### Language Discrete type

+ `language/type`: Description of this type of language subtype
  + `param (type)`: Description of optional parameters
+ `language/c++`: C++ Source Code
  + example: `language/c++; compiler=clang; features=O0:sanitize=asan,undefined`
  + `standard (enum)`: iso c++11, gnu c++17, etc.
  + `compiler (enum)`: a compiler-id with an optional [semver](https://semver.org/), clang 11.0, g++ 7.3, etc.
  + `features (string slice)`: seperated by colon, e.g. `O0:asan:-fsanitize=undefined`
+ `language/c`: see `language/c++`
+ `language/rust`
+ `language/python`
+ `language/java`
+ `language/golang`
+ `language/javascript`

#### Submit a precompiled binary

It is dangerous but the intra judge service should allow a super user to do this. Use `application/octec-stream` as the
identifier.
